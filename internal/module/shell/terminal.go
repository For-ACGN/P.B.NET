package shell

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"project/internal/convert"
	"project/internal/logger"
	"project/internal/system"
	"project/internal/xpanic"
)

// Welcome to terminal [version 1.0.0]
//
// C:\Windows\System32>

// Terminal is a interactive platform-independent system shell.
type Terminal struct {
	// input pipe
	iPr *os.File
	iPw *os.File

	// output pipe
	oPr *os.File
	oPw *os.File

	// multi writer, record user input
	input io.Writer

	// status
	cd    string
	env   []string
	cmd   *exec.Cmd
	cmdMu sync.Mutex

	closeOnce sync.Once
	wg        sync.WaitGroup
}

// NewTerminal is used to create a platform-independent system shell.
// copy is used to control input.
func NewTerminal(copy bool) (*Terminal, error) {
	cd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	terminal := Terminal{
		cd:  cd,
		env: os.Environ(),
	}
	iPr, iPw, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	terminal.iPr = iPr
	terminal.iPw = iPw
	oPr, oPw, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	terminal.oPr = oPr
	terminal.oPw = oPw
	if copy {
		terminal.input = io.MultiWriter(terminal.iPw, terminal.oPw)
	} else {
		terminal.input = terminal.iPw
	}
	terminal.wg.Add(1)
	go terminal.readInputLoop()
	return &terminal, nil
}

// Read is used to read terminal output data.
func (t *Terminal) Read(data []byte) (int, error) {
	return t.oPr.Read(data)
}

// Write is used to write user input data.
func (t *Terminal) Write(data []byte) (int, error) {
	return t.input.Write(data)
}

// Close is used to close terminal, if this terminal is running a program,
// it will be kill at the same time.
func (t *Terminal) Close() error {
	cmd := t.getProcess()
	if cmd != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Process.Release()
	}
	t.close()
	t.wg.Wait()
	return nil
}

// Interrupt is used to send interrupt signal to opened process.
func (t *Terminal) Interrupt() error {
	cmd := t.getProcess()
	if cmd != nil {
		return sendInterruptSignal(t.cmd)
	}
	return nil
}

func (t *Terminal) close() {
	t.closeOnce.Do(func() {
		_ = t.iPw.Close()
		_ = t.iPr.Close()
		_ = t.oPw.Close()
		_ = t.oPr.Close()
	})
}

// readInputLoop is used to read user input command and run.
func (t *Terminal) readInputLoop() {
	defer t.wg.Done()
	defer func() {
		if r := recover(); r != nil {
			buf := xpanic.Print(r, "Terminal.readInputLoop")
			_, _ = buf.WriteTo(t.oPw)
			// restart
			time.Sleep(time.Second)
			t.wg.Add(1)
			go t.readInputLoop()
		}
	}()
	// print welcome information
	hello := []byte("Welcome to terminal [version 1.0.0]\n\n")
	_, _ = t.oPw.Write(hello)
	t.printCurrentDirectory()
	scanner := bufio.NewScanner(t.iPr)
	for scanner.Scan() {
		input := scanner.Text()
		// check is internal command
		commandLine := strings.SplitN(trimPrefixSpace(input), " ", 2)
		command := commandLine[0]
		var args string
		if len(commandLine) == 2 {
			args = commandLine[1]
		}
		if t.executeInternalCommand(command, args) {
			t.printCurrentDirectory()
			continue
		}
		t.startProcess(input)
		t.printCurrentDirectory()
	}
}

func (t *Terminal) setProcess(cmd *exec.Cmd) {
	t.cmdMu.Lock()
	defer t.cmdMu.Unlock()
	t.cmd = cmd
}

func (t *Terminal) getProcess() *exec.Cmd {
	t.cmdMu.Lock()
	defer t.cmdMu.Unlock()
	return t.cmd
}

func (t *Terminal) startProcess(command string) bool {
	argv := system.CommandLineToArgv(command)
	cmd := exec.Command(argv[0], argv[1:]...) // #nosec
	cmd.SysProcAttr = setSysProcAttr()
	// program output
	cmd.Stdin = t.iPr
	cmd.Stderr = t.oPw
	cmd.Stdout = t.oPw
	cmd.Dir = t.cd
	// copy environment variable
	env := make([]string, len(t.env))
	copy(env, t.env)
	cmd.Env = env
	err := cmd.Start()
	if err != nil {
		_, _ = fmt.Fprintf(t.oPw, "%s\n\n", err)
		return false
	}
	t.setProcess(cmd)
	_ = cmd.Wait()
	_, _ = fmt.Fprintln(t.oPw)
	return true
}

// trimPrefixSpace is used to remove space in prefix
// "  a" -> "a"
func trimPrefixSpace(s string) string {
	r := []rune(s)
	l := len(r)
	space, _ := utf8.DecodeRuneInString(" ")
	for i := 0; i < l; i++ {
		if r[i] != space {
			return string(r[i:])
		}
	}
	return ""
}

func (t *Terminal) printCurrentDirectory() {
	line := []byte(t.cd + ">")
	_, _ = t.oPw.Write(line)
}

func (t *Terminal) executeInternalCommand(cmd, args string) bool {
	switch cmd {
	case "":
		// no input
	case "cd": // change current path
		t.changeDirectory(args)
	case "set": // set environment variable
		t.environmentVariable(args)
	case "dir", "ls":
		t.dir(args)
	case "exit":
		t.close()
	default:
		return false
	}
	return true
}

func (t *Terminal) changeDirectory(args string) {
	cd := system.CommandLineToArgv(args)
	if len(cd) == 0 {
		return
	}
	path := cd[0]
	var dstPath string
	if filepath.IsAbs(path) {
		dstPath = path
	} else {
		dstPath = t.cd + "/" + path
	}
	absPath, err := filepath.Abs(dstPath)
	if err != nil {
		_, _ = fmt.Fprintf(t.oPw, "%s\n\n", err)
		return
	}
	// check is exist
	_, err = os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			_, _ = fmt.Fprintf(t.oPw, "directory \"%s\" is not exist\n\n", path)
		} else {
			_, _ = fmt.Fprintf(t.oPw, "%s\n\n", err)
		}
		return
	}
	t.cd = absPath
	// print empty line
	_, _ = fmt.Fprintln(t.oPw)
}

func (t *Terminal) environmentVariable(args string) {
	args = trimPrefixSpace(args)
	if args == "" {
		t.printEnvironmentVariable()
		return
	}
	// has "="
	if !strings.Contains(args, "=") {
		t.findEnvironmentVariable(args)
		return
	}
	t.setEnvironmentVariable(args)
}

func (t *Terminal) printEnvironmentVariable() {
	buf := new(bytes.Buffer)
	for i := 0; i < len(t.env); i++ {
		_, _ = fmt.Fprintln(buf, t.env[i])
	}
	// print empty line
	_, _ = fmt.Fprintln(buf)
	_, _ = buf.WriteTo(t.oPw)
}

func (t *Terminal) findEnvironmentVariable(name string) {
	buf := new(bytes.Buffer)
	var find bool
	for i := 0; i < len(t.env); i++ {
		if strings.HasPrefix(strings.ToLower(t.env[i]), strings.ToLower(name)) {
			_, _ = fmt.Fprintln(buf, t.env[i])
			find = true
		}
	}
	if !find {
		const format = "environment variable %s is not defined\n"
		_, _ = fmt.Fprintf(buf, format, name)
	}
	// print empty line
	_, _ = fmt.Fprintln(buf)
	_, _ = buf.WriteTo(t.oPw)
}

func (t *Terminal) setEnvironmentVariable(args string) {
	nv := strings.SplitN(args, "=", 2)
	name := nv[0]
	value := nv[1]
	if name == "" {
		_, _ = fmt.Fprintf(t.oPw, "no variable name\n\n")
		return
	}
	if value == "" { // delete
		for i := 0; i < len(t.env); i++ {
			tNV := strings.SplitN(t.env[i], "=", 2)
			tName := tNV[0]
			if tName == name {
				t.env = append(t.env[:i], t.env[i+1:]...)
				break
			}
		}
	} else { // set or add
		var added bool
		for i := 0; i < len(t.env); i++ {
			tNV := strings.SplitN(t.env[i], "=", 2)
			tName := tNV[0]
			if tName == name {
				t.env[i] = args
				added = true
				break
			}
		}
		if !added {
			t.env = append(t.env, args)
		}
	}
	// print empty line
	_, _ = fmt.Fprintln(t.oPw)
}

func (t *Terminal) dir(args string) {
	// calculate path
	dir := system.CommandLineToArgv(args)
	var (
		rawPath string
		path    string
		list    []os.FileInfo
		err     error
	)
	if len(dir) == 0 {
		rawPath = t.cd
		path = t.cd
	} else {
		rawPath = dir[0]
		if filepath.IsAbs(rawPath) {
			path = rawPath
		} else {
			path = t.cd + "/" + rawPath
		}
	}
	list, err = ioutil.ReadDir(path)
	if err != nil {
		_, _ = fmt.Fprintf(t.oPw, "directory \"%s\" is not exist\n\n", rawPath)
		return
	}
	// sort, directory first, then file.
	sort.Slice(list, func(i, j int) bool {
		return strings.ToLower(list[i].Name()) < strings.ToLower(list[j].Name())
	})
	var (
		dirList  []os.FileInfo
		fileList []os.FileInfo
		// about format
		maxSizeStrLen int
		maxSizeLen    int
	)
	for i := 0; i < len(list); i++ {
		if list[i].IsDir() {
			dirList = append(dirList, list[i])
		} else {
			size := list[i].Size()
			sizeStrLen := len(convert.FormatByte(uint64(size)))
			sizeLen := len(convert.FormatNumber(strconv.FormatInt(size, 10)))
			if sizeStrLen > maxSizeStrLen {
				maxSizeStrLen = sizeStrLen
			}
			if sizeLen > maxSizeLen {
				maxSizeLen = sizeLen
			}
			fileList = append(fileList, list[i])
		}
	}
	// print
	const (
		splitSpaceSize = 1
		fixedSize      = 5 // " Byte"
	)
	paddingSize := maxSizeStrLen + splitSpaceSize + maxSizeLen
	if maxSizeStrLen > 0 {
		paddingSize += fixedSize
	}
	buf := new(bytes.Buffer)
	for i := 0; i < len(dirList); i++ {
		info := dirList[i]
		// time
		_, _ = fmt.Fprintf(buf, "%s ", info.ModTime().Format(logger.TimeLayout))
		// mode
		_, _ = fmt.Fprintf(buf, "%s ", info.Mode().Perm())
		// dir
		_, _ = fmt.Fprint(buf, "<dir> ")
		// size
		_, _ = fmt.Fprint(buf, strings.Repeat(" ", paddingSize))
		_, _ = fmt.Fprintf(buf, " %s\n", info.Name())
	}
	for i := 0; i < len(fileList); i++ {
		info := fileList[i]
		// time
		_, _ = fmt.Fprintf(buf, "%s ", info.ModTime().Format(logger.TimeLayout))
		// mode
		_, _ = fmt.Fprintf(buf, "%s ", info.Mode().Perm())
		// not dir
		_, _ = fmt.Fprint(buf, "      ")
		size := info.Size()
		format := "%" + strconv.Itoa(maxSizeStrLen) + "s"
		_, _ = fmt.Fprintf(buf, format, convert.FormatByte(uint64(size)))
		_, _ = fmt.Fprint(buf, strings.Repeat(" ", splitSpaceSize))
		format = "%" + strconv.Itoa(maxSizeLen) + "s Byte"
		_, _ = fmt.Fprintf(buf, format, convert.FormatNumber(strconv.FormatInt(size, 10)))
		_, _ = fmt.Fprintf(buf, " %s\n", info.Name())
	}
	_, _ = buf.WriteTo(t.oPw)
	// print empty line
	_, _ = fmt.Fprintln(t.oPw)
}
