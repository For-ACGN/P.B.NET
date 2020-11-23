package logger

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"project/internal/system"
	"project/internal/xpanic"
)

// Logger is used to print log with level and source.
type Logger interface {
	Printf(lv Level, src, format string, log ...interface{})
	Print(lv Level, src string, log ...interface{})
	Println(lv Level, src string, log ...interface{})
}

var (
	// Common is a common logger, some tools need it.
	Common Logger = new(common)

	// Test is used to go test.
	Test Logger = new(test)

	// Discard is used to discard log in object test.
	Discard Logger = new(discard)
)

// [2020-01-21 12:36:41] [info] <test src> test-format test log
type common struct{}

func (common) Printf(lv Level, src, format string, log ...interface{}) {
	if lv < Info {
		return
	}
	output := Prefix(time.Now(), lv, src)
	_, _ = fmt.Fprintf(output, format, log...)
	fmt.Println(output)
}

func (common) Print(lv Level, src string, log ...interface{}) {
	if lv < Info {
		return
	}
	output := Prefix(time.Now(), lv, src)
	_, _ = fmt.Fprint(output, log...)
	fmt.Println(output)
}

func (common) Println(lv Level, src string, log ...interface{}) {
	if lv < Info {
		return
	}
	output := Prefix(time.Now(), lv, src)
	_, _ = fmt.Fprintln(output, log...)
	fmt.Print(output)
}

// [Test] [2020-01-21 12:36:41] [debug] <test src> test-format test log
type test struct{}

var testLoggerPrefix = []byte("[Test] ")

func writePrefix(lv Level, src string) *bytes.Buffer {
	output := new(bytes.Buffer)
	output.Write(testLoggerPrefix)
	_, _ = Prefix(time.Now(), lv, src).WriteTo(output)
	return output
}

func (test) Printf(lv Level, src, format string, log ...interface{}) {
	output := writePrefix(lv, src)
	_, _ = fmt.Fprintf(output, format, log...)
	fmt.Println(output)
}

func (test) Print(lv Level, src string, log ...interface{}) {
	output := writePrefix(lv, src)
	_, _ = fmt.Fprint(output, log...)
	fmt.Println(output)
}

func (test) Println(lv Level, src string, log ...interface{}) {
	output := writePrefix(lv, src)
	_, _ = fmt.Fprintln(output, log...)
	fmt.Print(output)
}

type discard struct{}

func (discard) Printf(Level, string, string, ...interface{}) {}

func (discard) Print(Level, string, ...interface{}) {}

func (discard) Println(Level, string, ...interface{}) {}

// MultiLogger is a common logger that can set log level and print log.
type MultiLogger struct {
	writer io.Writer
	level  Level
	rwm    sync.RWMutex
}

// NewMultiLogger is used to create a MultiLogger.
func NewMultiLogger(lv Level, writers ...io.Writer) *MultiLogger {
	return &MultiLogger{
		level:  lv,
		writer: io.MultiWriter(writers...),
	}
}

// Printf is used to print log with format.
func (lg *MultiLogger) Printf(lv Level, src, format string, log ...interface{}) {
	lg.rwm.RLock()
	defer lg.rwm.RUnlock()
	if lv < lg.level {
		return
	}
	buf := Prefix(time.Now(), lv, src)
	_, _ = fmt.Fprintf(buf, format, log...)
	buf.WriteString("\n")
	_, _ = buf.WriteTo(lg.writer)
}

// Print is used to print log.
func (lg *MultiLogger) Print(lv Level, src string, log ...interface{}) {
	lg.rwm.RLock()
	defer lg.rwm.RUnlock()
	if lv < lg.level {
		return
	}
	buf := Prefix(time.Now(), lv, src)
	_, _ = fmt.Fprint(buf, log...)
	buf.WriteString("\n")
	_, _ = buf.WriteTo(lg.writer)
}

// Println is used to print log with new line.
func (lg *MultiLogger) Println(lv Level, src string, log ...interface{}) {
	lg.rwm.RLock()
	defer lg.rwm.RUnlock()
	if lv < lg.level {
		return
	}
	buf := Prefix(time.Now(), lv, src)
	_, _ = fmt.Fprintln(buf, log...)
	_, _ = buf.WriteTo(lg.writer)
}

// SetLevel is used to set log level that need print.
func (lg *MultiLogger) SetLevel(lv Level) error {
	if lv > Off {
		return fmt.Errorf("invalid logger level: %d", lv)
	}
	lg.rwm.Lock()
	defer lg.rwm.Unlock()
	lg.level = lv
	return nil
}

// Close is used to close logger.
func (lg *MultiLogger) Close() error {
	_ = lg.SetLevel(Off)
	return nil
}

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
	skip   int
	last   bool // reserve the last "\n"
}

func (w *wrapWriter) Write(p []byte) (int, error) {
	l := len(p)
	buf := bytes.NewBuffer(make([]byte, 0, l+256))
	buf.Write(p)
	if !w.last {
		buf.Truncate(buf.Len() - 1)
	}
	if w.trace {
		buf.WriteString("\n")
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
func SetErrorLogger(name string) (*os.File, error) {
	file, err := system.OpenFile(name, os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return nil, err
	}
	mLogger := NewMultiLogger(Warning, os.Stdout, file)
	HijackLogWriter(Fatal, "init", mLogger)
	return file, nil
}
