package logger

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"

	"project/internal/system"
	"project/internal/xpanic"
)

// prefixWriter is used to print with a prefix.
type prefixWriter struct {
	writer io.Writer
	prefix []byte
}

func (p *prefixWriter) Write(b []byte) (n int, err error) {
	n = len(b)
	_, err = p.writer.Write(append(p.prefix, b...))
	return
}

// NewWriterWithPrefix is used to print prefix before each log.
// It used to test role.
func NewWriterWithPrefix(w io.Writer, prefix string) io.Writer {
	return &prefixWriter{
		writer: w,
		prefix: []byte(fmt.Sprintf("[%s] ", prefix)),
	}
}

// wrapWriter will print stack trace to inner logger.
type wrapWriter struct {
	level  Level
	src    string
	logger Logger
	trace  bool // print stack trace
	skip   int  // about trace
	last   bool // delete the last "\n"
}

func (w *wrapWriter) Write(p []byte) (int, error) {
	l := len(p)
	buf := bytes.NewBuffer(make([]byte, 0, l+256))
	buf.Write(p)
	if w.last && p[len(p)-1] == '\n' {
		buf.Truncate(buf.Len() - 1)
	}
	if w.trace {
		xpanic.PrintStackTrace(buf, w.skip)
	}
	w.logger.Print(w.level, w.src, buf)
	return l, nil
}

// WrapLogger is used to wrap a Logger to io.Writer.
func WrapLogger(lv Level, src string, logger Logger) io.Writer {
	w := wrapWriter{
		level:  lv,
		src:    src,
		logger: logger,
		last:   true,
	}
	return &w
}

// Wrap is used to convert Logger to go internal logger.
// It used to set to http.Server.ErrorLog or other structure.
func Wrap(lv Level, src string, logger Logger) *log.Logger {
	w := wrapWriter{
		level:  lv,
		src:    src,
		logger: logger,
		trace:  true,
		skip:   3,
	}
	return log.New(&w, "", 0)
}

// HijackLogWriter is used to hijack all packages that call functions like log.Println().
func HijackLogWriter(lv Level, src string, logger Logger) {
	w := &wrapWriter{
		level:  lv,
		src:    src,
		logger: logger,
		trace:  true,
		skip:   4,
	}
	log.SetFlags(0)
	log.SetOutput(w)
}

// SetErrorLogger is used to log error before service program start.
// If occur some error before start, you can get it.
func SetErrorLogger(path string) (*os.File, error) {
	file, err := system.OpenFile(path, os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return nil, err
	}
	mLogger, _ := NewMultiLogger(Warning, os.Stdout, file)
	HijackLogWriter(Fatal, "init", mLogger)
	return file, nil
}
