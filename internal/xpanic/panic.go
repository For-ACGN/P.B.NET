package xpanic

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"runtime"
)

const maxDepth = 32

// PrintStack is used to print current stack to a *bytes.Buffer.
func PrintStack(buf *bytes.Buffer, skip int) {
	defer func() {
		if r := recover(); r != nil {
			buf.WriteString("\nfailed to print stack\n")
		}
	}()
	if skip > maxDepth {
		skip = 0
	}
	var pcs [maxDepth]uintptr
	n := runtime.Callers(skip, pcs[:])
	// skip unnecessary pc
	for _, pc := range pcs[2 : n-1] {
		f := frame(pc)
		// write source file
		pc := f.pc()
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			_, _ = buf.WriteString("unknown")
		} else {
			file, _ := fn.FileLine(pc)
			_, _ = fmt.Fprintf(buf, "%s\n\t%s", fn.Name(), file)
		}
		_, _ = buf.WriteString(":")
		_, _ = fmt.Fprintf(buf, "%d\n", f.line())
	}
	// remove the last new line
	if buf.Len() > 1 {
		buf.Truncate(buf.Len() - 1)
	}
}

// frame represents a program counter inside a stack frame.
type frame uintptr

// pc returns the program counter for this frame;
// multiple frames may have the same PC value.
func (f frame) pc() uintptr {
	return uintptr(f) - 1
}

// line returns the line number of source code of the
// function for this Frame's pc.
func (f frame) line() int {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return 0
	}
	_, line := fn.FileLine(f.pc())
	return line
}

// PrintStackTrace is used to print stack with line.
func PrintStackTrace(buf *bytes.Buffer, skip int) {
	buf.WriteString("--------------------------stack trace---------------------------\n")
	PrintStack(buf, skip)
	buf.WriteString("\n----------------------------------------------------------------")
}

// PrintPanic is used to print panic to a *bytes.Buffer.
func PrintPanic(panic interface{}, title string, skip int) *bytes.Buffer {
	buf := new(bytes.Buffer)
	buf.WriteString(title)
	buf.WriteString(":\n")
	_, _ = fmt.Fprintln(buf, panic)
	PrintStackTrace(buf, skip)
	return buf
}

// Print is used to print panic and stack to a *bytes.Buffer.
func Print(panic interface{}, title string) *bytes.Buffer {
	return PrintPanic(panic, title, 3)
}

// Printf is used to printf panic and stack to a *bytes.Buffer.
func Printf(panic interface{}, format string, args ...interface{}) *bytes.Buffer {
	return PrintPanic(panic, fmt.Sprintf(format, args...), 3)
}

// Error is used to print panic and stack to a *bytes.Buffer buf and return an error.
func Error(panic interface{}, title string) error {
	return errors.New(Print(panic, title).String())
}

// Errorf is used to printf panic and stack to a *bytes.Buffer buf and return an error.
func Errorf(panic interface{}, format string, args ...interface{}) error {
	return errors.New(Printf(panic, format, args...).String())
}

// Log is used to call log.Println to print panic and stack.
// It is used to log in some package without logger.Logger.
func Log(panic interface{}, title string) *bytes.Buffer {
	buf := Print(panic, title)
	log.Println(buf)
	return buf
}

// Logf is used to call log.Println to printf panic and stack.
// It is used to log in some package without logger.Logger.
func Logf(panic interface{}, format string, args ...interface{}) *bytes.Buffer {
	buf := Printf(panic, format, args...)
	log.Println(buf)
	return buf
}
