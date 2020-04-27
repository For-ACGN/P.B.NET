package logger

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

// Level is the log level.
type Level = uint8

// about level
const (
	Debug Level = iota
	Info
	Warning
	Error
	Exploit
	Fatal
	Off
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
	lv := Level(0)
	switch strings.ToLower(level) {
	case "debug":
		lv = Debug
	case "info":
		lv = Info
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
// [2018-11-27 00:00:00] [info] <socks5-test> start listener
func Prefix(time time.Time, level Level, src string) *bytes.Buffer {
	var lv string
	switch level {
	case Debug:
		lv = "debug"
	case Info:
		lv = "info"
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

// [2020-01-21 12:36:41] [debug] <test src> test-format test log
type common struct{}

func (common) Printf(lv Level, src, format string, log ...interface{}) {
	output := Prefix(time.Now(), lv, src)
	_, _ = fmt.Fprintf(output, format, log...)
	fmt.Println(output)
}

func (common) Print(lv Level, src string, log ...interface{}) {
	output := Prefix(time.Now(), lv, src)
	_, _ = fmt.Fprint(output, log...)
	fmt.Println(output)
}

func (common) Println(lv Level, src string, log ...interface{}) {
	output := Prefix(time.Now(), lv, src)
	_, _ = fmt.Fprintln(output, log...)
	fmt.Print(output)
}

// [Test] [2020-01-21 12:36:41] [debug] <test src> test-format test log
type test struct{}

var testPrefix = []byte("[Test] ")

func writePrefix(lv Level, src string) *bytes.Buffer {
	output := new(bytes.Buffer)
	output.Write(testPrefix)
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
// it used for role test
func NewWriterWithPrefix(w io.Writer, prefix string) io.Writer {
	return &prefixWriter{
		writer: w,
		prefix: []byte(fmt.Sprintf("[%s] ", prefix)),
	}
}

type wrapWriter struct {
	level  Level
	src    string
	logger Logger
}

func (w *wrapWriter) Write(p []byte) (int, error) {
	w.logger.Println(w.level, w.src, string(p[:len(p)-1]))
	return len(p), nil
}

// Wrap is used to convert Logger to go internal logger like http.Server.ErrorLog.
func Wrap(lv Level, src string, logger Logger) *log.Logger {
	w := &wrapWriter{
		level:  lv,
		src:    src,
		logger: logger,
	}
	return log.New(w, "", 0)
}

// HijackLogWriter is used to hijack all packages that use log.Print().
func HijackLogWriter(lv Level, src string, logger Logger, flag int) {
	w := &wrapWriter{
		level:  lv,
		src:    src,
		logger: logger,
	}
	log.SetFlags(flag)
	log.SetOutput(w)
}

// Conn is used to print connection information.
// local:  tcp 127.0.0.1:123
// remote: tcp 127.0.0.1:124
func Conn(conn net.Conn) *bytes.Buffer {
	b := bytes.Buffer{}
	_, _ = fmt.Fprintf(&b, "local:  %s %s\nremote: %s %s ",
		conn.LocalAddr().Network(), conn.LocalAddr(),
		conn.RemoteAddr().Network(), conn.RemoteAddr())
	return &b
}
