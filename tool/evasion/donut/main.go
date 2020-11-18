package main

import (
	"bytes"
	"compress/flate"
	"debug/pe"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"text/template"

	"project/external/go-donut/donut"

	"project/internal/convert"
	"project/internal/crypto/aes"
	"project/internal/module/shellcode"
	"project/internal/random"
	"project/internal/system"
)

func main() {
	var (
		input   string
		output  string
		entropy uint64
		params  string
		thread  bool
		noGUI   bool
		method  string
		scOnly  bool
	)
	usage := "input executable file path"
	flag.StringVar(&input, "i", "", usage)
	usage = "output executable file path"
	flag.StringVar(&output, "o", "output.exe", usage)
	usage = "command line for exe"
	flag.StringVar(&params, "p", "", usage)
	usage = "create new thread, see document about donut"
	flag.BoolVar(&thread, "thread", false, usage)
	usage = "hide Windows GUI"
	flag.BoolVar(&noGUI, "no-gui", false, usage)
	usage = "entropy, see document about donut"
	flag.Uint64Var(&entropy, "entropy", donut.EntropyNone, usage)
	usage = "shellcode execute method"
	flag.StringVar(&method, "method", shellcode.MethodVirtualProtect, usage)
	usage = "only generate shellcode"
	flag.BoolVar(&scOnly, "sc", false, usage)
	flag.Parse()

	// load executable file
	if input == "" {
		fmt.Println("no input file path")
		return
	}
	exeData, err := ioutil.ReadFile(input) // #nosec
	system.CheckError(err)

	// set donut configuration
	donutCfg := donut.DefaultConfig()
	donutCfg.ExitOpt = 1 // exit thread
	if thread {
		donutCfg.Thread = 1 // create new thread
	}

	// read architecture
	peFile, err := pe.NewFile(bytes.NewReader(exeData))
	system.CheckError(err)
	var arch string
	switch peFile.Machine {
	case pe.IMAGE_FILE_MACHINE_I386:
		arch = "386"
		donutCfg.Arch = donut.X32
	case pe.IMAGE_FILE_MACHINE_AMD64:
		arch = "amd64"
		donutCfg.Arch = donut.X64
	default:
		fmt.Printf("unsupported executable file: 0x%02X\n", peFile.Machine)
		return
	}
	fmt.Println("[info] the architecture of the executable file is", arch)

	// convert to shellcode
	donutCfg.Entropy = uint32(entropy)
	donutCfg.Parameters = params
	scBuf, err := donut.ShellcodeFromBytes(bytes.NewBuffer(exeData), donutCfg)
	system.CheckError(err)
	fmt.Println("[info] convert executable file to shellcode")

	// save shellcode
	if scOnly {
		if output == "output.exe" {
			output = "output.bin"
		}
		err = system.WriteFile(output, scBuf.Bytes())
		system.CheckError(err)
		fmt.Println("[info] save shellcode")
		return
	}

	// compress shellcode
	fmt.Println("[info] compress generated shellcode")
	flateBuf := bytes.NewBuffer(make([]byte, 0, scBuf.Len()/2))
	writer, err := flate.NewWriter(flateBuf, flate.BestCompression)
	system.CheckError(err)
	_, err = scBuf.WriteTo(writer)
	system.CheckError(err)
	err = writer.Close()
	system.CheckError(err)

	// encrypt shellcode
	fmt.Println("[info] encrypt compressed shellcode")
	aesKey := random.Bytes(aes.Key256Bit)
	aesIV := random.Bytes(aes.IVSize)
	encShellcode, err := aes.CBCEncrypt(flateBuf.Bytes(), aesKey, aesIV)
	system.CheckError(err)

	// generate source code
	fmt.Println("[info] generate source code")
	tpl := template.New("execute")
	_, err = tpl.Parse(srcTemplate)
	system.CheckError(err)
	const tempSrc = "temp.go"
	srcFile, err := os.Create(tempSrc)
	system.CheckError(err)
	defer func() {
		_ = srcFile.Close()
		_ = os.Remove(tempSrc)
	}()
	cfg := config{
		Shellcode: convert.OutputBytes(encShellcode),
		AESKey:    convert.OutputBytes(aesKey),
		AESIV:     convert.OutputBytes(aesIV),
		Method:    method,
	}
	err = tpl.Execute(srcFile, cfg)
	system.CheckError(err)

	// build source code
	fmt.Println("[info] build source code to final executable file")
	ldFlags := "-s -w"
	if noGUI {
		ldFlags += " -H windowsgui"
	}
	args := []string{"build", "-v", "-trimpath", "-ldflags", ldFlags, "-o", output, tempSrc}
	cmd := exec.Command("go", args...) // #nosec
	cmd.Env = append(os.Environ(), "GOOS=windows")
	cmd.Env = append(cmd.Env, "GOARCH="+arch)
	cmdOutput, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(cmdOutput))
		fmt.Println(err)
		return
	}
	fmt.Println("[info] build final executable file successfully")
}

type config struct {
	Shellcode string
	AESKey    string
	AESIV     string
	Method    string
}

const srcTemplate = `
package main

import (
	"bytes"
	"compress/flate"
	"fmt"

	"project/internal/crypto/aes"
	"project/internal/module/shellcode"
)

func main() {
	encShellcode := {{.Shellcode}}

	// decrypt shellcode
	aesKey := {{.AESKey}}
	aesIV := {{.AESIV}}
	decShellcode, err := aes.CBCDecrypt(encShellcode, aesKey, aesIV)
	if err != nil {
		fmt.Println("failed to decrypt shellcode:", err)
		return
	}

	// decompress shellcode
	rc := flate.NewReader(bytes.NewReader(decShellcode))
	sc := bytes.NewBuffer(make([]byte, 0, len(decShellcode)*2))
	_, err = sc.ReadFrom(rc)
	if err != nil {
		fmt.Println("failed to decompress shellcode:", err)
		return
	}

	// execute shellcode
	method := "{{.Method}}"
	err = shellcode.Execute(method, sc.Bytes())
	if err != nil {
		fmt.Println("failed to execute shellcode:", err)
	}
}
`
