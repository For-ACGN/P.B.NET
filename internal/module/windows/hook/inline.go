// +build windows

package hook

import (
	"bytes"
	"encoding/binary"
	"reflect"
	"runtime"
	"sync"
	"unsafe"

	"github.com/pkg/errors"
	"golang.org/x/arch/x86/x86asm"
	"golang.org/x/sys/windows"

	"project/internal/module/windows/api"
	"project/internal/random"
)

// if want to develop, build program that use this package like rdpthief,
// must use "project/build.bat[sh] -install-hook" before build and
// use "project/build.bat[sh] -uninstall-hook" to restore runtime files,
// otherwise maybe appear unexpected problems.

func init() {
	runtime.InitASMStdcallAddr()
}

// [patch] is a near jump to [hook jumper].
// [hook jumper] is a far jump to our hook function.
// [trampoline] is a part of code about original function and
// add a near jump to the remaining original function.

// <security> not use FlushInstructionCache for bypass AV.

// PatchGuard contain information about hooked function.
type PatchGuard struct {
	Original *windows.Proc

	// contain origin data before hook
	fnData    []byte
	patchMem  *memory
	patchData []byte

	// about hook jumper
	hookJumperMem  *memory
	hookJumperData []byte

	// about trampoline function
	trampolineMem  *memory
	trampolineData []byte

	mu sync.Mutex
}

// Patch is used to patch the target function.
func (pg *PatchGuard) Patch() error {
	pg.mu.Lock()
	defer pg.mu.Unlock()
	// must create hook jumper before patch
	err := pg.hookJumperMem.Write(pg.hookJumperData)
	if err != nil {
		return errors.WithMessage(err, "failed to create hook jumper")
	}
	err = pg.trampolineMem.Write(pg.trampolineData)
	if err != nil {
		return errors.WithMessage(err, "failed to create trampoline function")
	}
	err = pg.patchMem.Write(pg.patchData)
	if err != nil {
		return errors.WithMessage(err, "failed to create patch")
	}
	return nil
}

// UnPatch is used to unpatch the target function.
func (pg *PatchGuard) UnPatch() error {
	pg.mu.Lock()
	defer pg.mu.Unlock()
	// must unload patch before recover hook jumper
	err := pg.patchMem.Write(pg.fnData)
	if err != nil {
		return errors.WithMessage(err, "failed to unload patch")
	}
	err = pg.hookJumperMem.Write(bytes.Repeat([]byte{0xCC}, len(pg.hookJumperData)))
	if err != nil {
		return errors.WithMessage(err, "failed to recover memory about hook jumper")
	}
	err = pg.trampolineMem.Write(bytes.Repeat([]byte{0xCC}, len(pg.trampolineData)))
	if err != nil {
		return errors.WithMessage(err, "failed to recover memory about trampoline function")
	}
	return nil
}

// NewInlineHookByName is used to create a hook about function by DLL name and Proc name.
func NewInlineHookByName(dll, proc string, system bool, hookFn interface{}) (*PatchGuard, error) {
	var lazyDLL *windows.LazyDLL
	if system {
		lazyDLL = windows.NewLazySystemDLL(dll)
	} else {
		lazyDLL = windows.NewLazyDLL(dll)
	}
	err := lazyDLL.Load()
	if err != nil {
		return nil, err
	}
	// find proc
	lazyProc := lazyDLL.NewProc(proc)
	err = lazyProc.Find()
	if err != nil {
		return nil, err
	}
	target := &windows.Proc{
		Name: dll,
	}
	target.Dll = &windows.DLL{
		Name:   dll,
		Handle: windows.Handle(lazyDLL.Handle()),
	}
	// set private structure field "addr"
	*(*uintptr)(unsafe.Pointer(
		reflect.ValueOf(target).Elem().FieldByName("addr").UnsafeAddr()),
	) = lazyProc.Addr() // #nosec
	return NewInlineHook(target, hookFn)
}

// NewInlineHook is used to create a hook about function, usually hook a syscall.
func NewInlineHook(target *windows.Proc, hookFn interface{}) (*PatchGuard, error) {
	// select architecture
	arch := newArch()
	// read function address
	targetAddr := target.Addr()
	hookFnAddr := windows.NewCallback(hookFn)
	// fmt.Printf("0x%X,0x%X\n", targetAddr, hookFnAddr)
	// inaccurate but sufficient
	// Prefix = 4, OpCode = 3, ModRM = 1, sib = 1, displacement = 4, immediate = 4
	const maxInstLen = 2 * (4 + 3 + 1 + 1 + 4 + 4)
	originalFunc, err := readMemory(targetAddr, maxInstLen)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read memory about original function")
	}
	// get instructions
	insts, err := disassemble(originalFunc, arch.DisassembleMode())
	if err != nil {
		return nil, errors.WithMessage(err, "failed to disassemble original function")
	}
	// get patch size that need fix for trampoline function
	patchSize, instNum, err := getASMPatchSizeAndInstNum(insts)
	if err != nil {
		return nil, err
	}
	// replace special instruction
	rebuiltInsts, newInsts, err := rebuildInstruction(originalFunc[:patchSize], insts[:instNum])
	if err != nil {
		return nil, err
	}
	// query memory information about target procedure address for
	// search writeable memory about hook jumper and trampoline.
	mbi, err := api.VirtualQuery(targetAddr)
	if err != nil {
		return nil, err
	}
	// create hook jumper
	hookJumperMem, err := searchMemory(targetAddr, arch.FarJumpSize(), true, mbi)
	if err != nil {
		return nil, err
	}
	hookJumperData := arch.NewFarJumpASM(hookJumperMem.Addr, hookFnAddr)
	// add padding data for prevent search the same memory
	err = hookJumperMem.Write(bytes.Repeat([]byte{0x00}, len(hookJumperData)))
	if err != nil {
		return nil, err
	}
	// fmt.Printf("hook jumper: 0x%X\n", hookJumperMem.Addr)
	// create patch for jump to hook jumper
	patchMem := newMemory(targetAddr, nearJumperSize)
	patchData := newNearJumpASM(targetAddr, hookJumperMem.Addr)
	// create trampoline function for call original function
	trampMem, err := searchMemory(targetAddr, len(rebuiltInsts)+nearJumperSize, false, mbi)
	if err != nil {
		return nil, err
	}
	tramData := relocateInstruction(int(trampMem.Addr-targetAddr), rebuiltInsts, newInsts)
	finalJump := newNearJumpASM(trampMem.Addr+uintptr(len(tramData)), targetAddr+uintptr(patchSize))
	tramData = append(tramData, finalJump...)
	// create proc for call original function
	fakeOriginalProc := &windows.Proc{
		Dll:  target.Dll,
		Name: target.Name,
	}
	*(*uintptr)(unsafe.Pointer(
		reflect.ValueOf(fakeOriginalProc).Elem().FieldByName("addr").UnsafeAddr()),
	) = trampMem.Addr // #nosec
	// create patch guard
	pg := PatchGuard{
		Original:       fakeOriginalProc,
		fnData:         originalFunc[:nearJumperSize],
		patchMem:       patchMem,
		patchData:      patchData,
		hookJumperMem:  hookJumperMem,
		hookJumperData: hookJumperData,
		trampolineMem:  trampMem,
		trampolineData: tramData,
	}
	err = pg.Patch()
	if err != nil {
		return nil, err
	}
	return &pg, nil
}

func disassemble(src []byte, mode int) ([]*x86asm.Inst, error) {
	var insts []*x86asm.Inst
	for len(src) > 0 {
		inst, err := x86asm.Decode(src, mode)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		insts = append(insts, &inst)
		src = src[inst.Len:]
	}
	return insts, nil
}

// if appear too short function, we can improve it that add jumper to it,
// and use behind address to storage old code.
func getASMPatchSizeAndInstNum(insts []*x86asm.Inst) (int, int, error) {
	var (
		size int
		num  int
	)
	for i := 0; i < len(insts); i++ {
		size += insts[i].Len
		num++
		if size >= nearJumperSize {
			return size, num, nil
		}
	}
	return 0, 0, errors.New("unable to insert near jmp to this function")
}

// searchMemory is used to search memory for write hook jumper and trampoline.
// if low is true search low address first.
func searchMemory(begin uintptr, size int, low bool, mbi *api.MemoryBasicInformation) (*memory, error) {
	mem, err := searchMemoryAt(begin, size, !low, mbi)
	if err == nil {
		return mem, nil
	}
	return searchMemoryAt(begin, size, low, mbi)
}

func searchMemoryAt(begin uintptr, size int, add bool, mbi *api.MemoryBasicInformation) (*memory, error) {
	// set random begin address
	var maxRange uintptr
	rand := random.NewRand()
	if add {
		boundary := mbi.AllocationBase + mbi.RegionSize
		begin += 16 + uintptr(rand.Int(int(boundary-begin)/10))
		maxRange = boundary - begin - uintptr(size)
	} else {
		boundary := mbi.AllocationBase
		begin -= uintptr(size) + uintptr(rand.Int(int(begin-boundary)/10))
		maxRange = begin - boundary
	}
	// search memory address
	var addr uintptr
	for i := uintptr(0); i < maxRange; i++ {
		if add {
			addr = begin + i
		} else {
			addr = begin - i
		}
		mem, err := readMemory(addr, size)
		if err != nil {
			return nil, err
		}
		// check memory is all int3 code
		if bytes.Equal(mem, bytes.Repeat([]byte{0xCC}, size)) {
			return newMemory(addr, size), nil
		}
	}
	return nil, errors.New("failed to search writeable memory")
}

// rebuildInstruction is used to rebuild instruction for replace jmp short to jmp ...
// [Warning]: it is only partially done.
func rebuildInstruction(src []byte, insts []*x86asm.Inst) ([]byte, []*x86asm.Inst, error) {
	rebuilt := make([]byte, 0, len(src))
	for i := 0; i < len(insts); i++ {
		switch insts[i].Op {
		case x86asm.JMP:
			switch src[0] {
			case 0xEB: // replace to near jump
				inst := make([]byte, 5)
				inst[0] = 0xE9
				// not think 1.
				mem := insts[i].Args[0].(x86asm.Mem)
				binary.LittleEndian.PutUint32(inst[1:], uint32(mem.Disp+3))
				rebuilt = append(rebuilt, inst...)
			default:
				rebuilt = append(rebuilt, src[:insts[i].Len]...)
			}
		default:
			rebuilt = append(rebuilt, src[:insts[i].Len]...)
		}
		src = src[insts[i].Len:]
	}
	newInsts, err := disassemble(rebuilt, insts[0].Mode)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "failed to disassemble rebuilt instruction")
	}
	return rebuilt, newInsts, nil
}

// relocateInstruction is used to relocate instruction like jmp, call...
// [Warning]: it is only partially done.
func relocateInstruction(offset int, src []byte, insts []*x86asm.Inst) []byte {
	relocated := make([]byte, 0, len(src))
	for i := 0; i < len(insts); i++ {
		switch insts[i].Op {
		case x86asm.CALL:
			switch src[0] {
			case 0xFF:
				switch src[1] {
				case 0x15:
					mem := insts[i].Args[0].(x86asm.Mem)
					binary.LittleEndian.PutUint32(src[2:], uint32(int(mem.Disp)-offset))
				}
			case 0xE8:
				mem := insts[i].Args[0].(x86asm.Mem)
				binary.LittleEndian.PutUint32(src[1:], uint32(int(mem.Disp)-offset))
			}
		case x86asm.JMP:
			switch src[0] {
			case 0xE9: // not think 1.
				mem := insts[i].Args[0].(x86asm.Mem)
				binary.LittleEndian.PutUint32(src[1:], uint32(int(mem.Disp)-offset))
			case 0x48:
				switch src[1] {
				case 0xFF:
					switch src[2] {
					case 0x25:
						mem := insts[i].Args[0].(x86asm.Mem)
						binary.LittleEndian.PutUint32(src[3:], uint32(int(mem.Disp)-offset))
					}
				}
			}
		}
		relocated = append(relocated, src[:insts[i].Len]...)
		src = src[insts[i].Len:]

		// spew.Config.DisableMethods = true
		// spew.Dump(insts[i])
	}
	return relocated
}
