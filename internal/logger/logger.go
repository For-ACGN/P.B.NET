package logger

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"project/internal/system"
	"project/internal/xpanic"
)

// Level is the log level.
type Level = uint8

// levels about logger.
const (
	All Level = iota // show all log messages

	// about debug
	Trace // for trace function (development)
	Debug // general debug information

	// about information
	Info    // common running information
	Crucial // important information like attack successfully

	// about error
	Warning // appear error but can continue
	Error   // appear error that can not continue (returned)
	Exploit // find attack exploit or security problem(maybe)
	Fatal   // appear panic in goroutine

	Off // stop log message
)

// TimeLayout is used to provide a parameter to time.Time.Format().
const TimeLayout = "2006-01-02 15:04:05"

// Logger is used to print log with level and source.
type Logger interface {
	Printf(lv Level, src, format string, log ...interface{})
	Print(lv Level, src string, log ...interface{})
	Println(lv Level, src string, log ...interface{})
}

// LevelSetter is used to set logger level.
type LevelSetter interface {
	SetLevel(lv Level) error
}

// Parse is used to parse logger level from string.
func Parse(level string) (Level, error) {
	var lv Level
	switch strings.ToLower(level) {
	case "all":
		lv = All
	case "trace":
		lv = Trace
	case "debug":
		lv = Debug
	case "info":
		lv = Info
	case "crucial":
		lv = Crucial
	case "warning":
		lv = Warning
	case "error":
		lv = Error
	case "exploit":
		lv = Exploit
	case "fatal":
		lv = Fatal
	case "off":
		lv = Off
	default:
		return lv, fmt.Errorf("unknown logger level: %s", level)
	}
	return lv, nil
}

// Prefix is used to print time, level and source to a buffer.
//
// time + level + source + log
// source usually like: class name + "-" + instance tag
//
// [2018-11-27 00:00:00] [info] <main> controller is running
// [2018-11-27 00:00:00] [info] <socks5-test> test log
func Prefix(time time.Time, level Level, src string) *bytes.Buffer {
	var lv string
	switch level {
	case Trace:
		lv = "trace"
	case Debug:
		lv = "debug"
	case Info:
		lv = "info"
	case Crucial:
		lv = "crucial"
	case Warning:
		lv = "warning"
	case Error:
		lv = "error"
	case Exploit:
		lv = "exploit"
	case Fatal:
		lv = "fatal"
	default:
		lv = "unknown"
	}
	buf := bytes.Buffer{}
	buf.WriteString("[")
	buf.WriteString(time.Local().Format(TimeLayout))
	buf.WriteString("] [")
	buf.WriteString(lv)
	buf.WriteString("] <")
	buf.WriteString(src)
	buf.WriteString("> ")
	return &buf
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

func (discard) Printf(_ Level, _, _ string, _ ...interface{}) {}

func (discard) Print(_ Level, _ string, _ ...interface{}) {}

func (discard) Println(_ Level, _ string, _ ...interface{}) {}

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

// Conn is used to print connection information to a *bytes.Buffer.
//
// local:  tcp 127.0.0.1:1234
// remote: tcp 127.0.0.1:1235
func Conn(conn net.Conn) *bytes.Buffer {
	buf := bytes.NewBuffer(make([]byte, 0, 64))
	_, _ = fmt.Fprintf(buf, "local:  %s %s\nremote: %s %s ",
		conn.LocalAddr().Network(), conn.LocalAddr(),
		conn.RemoteAddr().Network(), conn.RemoteAddr(),
	)
	return buf
}
