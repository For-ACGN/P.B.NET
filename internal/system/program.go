package system

import (
	"fmt"
	"os"
	"path/filepath"
)

// ExecutableName is used to get the executable file name.
func ExecutableName() (string, error) {
	path, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Base(path), nil
}

// ChangeCurrentDirectory is used to changed path for service program
// and prevent to get invalid path when running test.
func ChangeCurrentDirectory() error {
	path, err := os.Executable()
	if err != nil {
		return err
	}
	return os.Chdir(filepath.Dir(path))
}

// CheckError is used to check error is nil, if err is not nil,
// it will print error and exit program with code 1.
func CheckError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// PrintError is used to print error and exit program with code 1.
func PrintError(err ...interface{}) {
	fmt.Println(err...)
	os.Exit(1)
}

// CommandLineToArgv splits a command line into individual argument
// strings, following the Windows conventions documented
// at http://daviddeley.com/autohotkey/parameters/parameters.htm#WINARGV
func CommandLineToArgv(cmd string) []string {
	var args []string
	for len(cmd) > 0 {
		if cmd[0] == ' ' || cmd[0] == '\t' {
			cmd = cmd[1:]
			continue
		}
		var arg []byte
		arg, cmd = readNextArg(cmd)
		args = append(args, string(arg))
	}
	return args
}

// appendBSBytes appends n '\\' bytes to b and returns the resulting slice.
func appendBSBytes(b []byte, n int) []byte {
	for ; n > 0; n-- {
		b = append(b, '\\')
	}
	return b
}

// readNextArg splits command line string cmd into next
// argument and command line remainder.
func readNextArg(cmd string) (arg []byte, rest string) {
	var b []byte
	var inQuote bool
	var nSlash int
	for ; len(cmd) > 0; cmd = cmd[1:] {
		c := cmd[0]
		switch c {
		case ' ', '\t':
			if !inQuote {
				return appendBSBytes(b, nSlash), cmd[1:]
			}
		case '"':
			b = appendBSBytes(b, nSlash/2)
			if nSlash%2 == 0 {
				// use "Prior to 2008" rule from
				// http://daviddeley.com/autohotkey/parameters/parameters.htm
				// section 5.2 to deal with double double quotes
				if inQuote && len(cmd) > 1 && cmd[1] == '"' {
					b = append(b, c)
					cmd = cmd[1:]
				}
				inQuote = !inQuote
			} else {
				b = append(b, c)
			}
			nSlash = 0
			continue
		case '\\':
			nSlash++
			continue
		}
		b = appendBSBytes(b, nSlash)
		nSlash = 0
		b = append(b, c)
	}
	return appendBSBytes(b, nSlash), ""
}
