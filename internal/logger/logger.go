package logger

import (
	"bytes"
	"fmt"
	"io"
	"sync"
	"time"
)

// Logger is used to print log with level and source. If log
// level is lower than current level, this log will be discard.
type Logger interface {
	// Printf is used to log with format, reference features about fmt.Printf.
	Printf(lv Level, src, format string, log ...interface{})

	// Print is used to log with new line, reference features about fmt.Print.
	Print(lv Level, src string, log ...interface{})

	// Println is used to log with new line, reference features about fmt.Println.
	Println(lv Level, src string, log ...interface{})

	// SetLevel is used to set logger minimum log level.
	SetLevel(lv Level) error

	// GetLevel is used to get logger current minimum log level.
	GetLevel() Level
}

var (
	// Common is a common logger, most tools need it.
	Common, _ = NewCommonLogger(Info)

	// Test is used for go unit tests.
	Test, _ = NewTestLogger(Debug)

	// Discard is used to discard all log.
	Discard Logger = new(discard)
)

// [2020-01-21 12:36:41 +08:00] [info] <test src> test-format test log
type common struct {
	level Level
	rwm   sync.RWMutex
}

// NewCommonLogger is used to create a common logger.
func NewCommonLogger(lv Level) (Logger, error) {
	lg := common{}
	err := lg.SetLevel(lv)
	if err != nil {
		return nil, err
	}
	return &lg, nil
}

func (c *common) Printf(lv Level, src, format string, log ...interface{}) {
	if c.discard(lv) {
		return
	}
	output := Prefix(time.Now(), lv, src)
	_, _ = fmt.Fprintf(output, format, log...)
	fmt.Println(output)
}

func (c *common) Print(lv Level, src string, log ...interface{}) {
	if c.discard(lv) {
		return
	}
	output := Prefix(time.Now(), lv, src)
	_, _ = fmt.Fprint(output, log...)
	fmt.Println(output)
}

func (c *common) Println(lv Level, src string, log ...interface{}) {
	if c.discard(lv) {
		return
	}
	output := Prefix(time.Now(), lv, src)
	_, _ = fmt.Fprintln(output, log...)
	fmt.Print(output)
}

func (c *common) discard(lv Level) bool {
	c.rwm.RLock()
	defer c.rwm.RUnlock()
	return lv < c.level
}

func (c *common) SetLevel(lv Level) error {
	if lv > Off {
		return fmt.Errorf("invalid logger level: %d", lv)
	}
	c.rwm.Lock()
	defer c.rwm.Unlock()
	c.level = lv
	return nil
}

func (c *common) GetLevel() Level {
	c.rwm.RLock()
	defer c.rwm.RUnlock()
	return c.level
}

// [Test] [2020-01-21 12:36:41 +08:00] [debug] <test src> test-format test log
type test struct {
	level Level
	rwm   sync.RWMutex
}

// NewTestLogger is used to create a test logger.
func NewTestLogger(lv Level) (Logger, error) {
	lg := test{}
	err := lg.SetLevel(lv)
	if err != nil {
		return nil, err
	}
	return &lg, nil
}

func (t *test) Printf(lv Level, src, format string, log ...interface{}) {
	if t.discard(lv) {
		return
	}
	output := writePrefix(lv, src)
	_, _ = fmt.Fprintf(output, format, log...)
	fmt.Println(output)
}

func (t *test) Print(lv Level, src string, log ...interface{}) {
	if t.discard(lv) {
		return
	}
	output := writePrefix(lv, src)
	_, _ = fmt.Fprint(output, log...)
	fmt.Println(output)
}

func (t *test) Println(lv Level, src string, log ...interface{}) {
	if t.discard(lv) {
		return
	}
	output := writePrefix(lv, src)
	_, _ = fmt.Fprintln(output, log...)
	fmt.Print(output)
}

func (t *test) discard(lv Level) bool {
	t.rwm.RLock()
	defer t.rwm.RUnlock()
	return lv < t.level
}

func writePrefix(lv Level, src string) *bytes.Buffer {
	output := new(bytes.Buffer)
	output.WriteString("[Test] ")
	_, _ = Prefix(time.Now(), lv, src).WriteTo(output)
	return output
}

func (t *test) SetLevel(lv Level) error {
	if lv > Off {
		return fmt.Errorf("invalid logger level: %d", lv)
	}
	t.rwm.Lock()
	defer t.rwm.Unlock()
	t.level = lv
	return nil
}

func (t *test) GetLevel() Level {
	t.rwm.RLock()
	defer t.rwm.RUnlock()
	return t.level
}

type discard struct{}

func (discard) Printf(Level, string, string, ...interface{}) {}

func (discard) Print(Level, string, ...interface{}) {}

func (discard) Println(Level, string, ...interface{}) {}

func (discard) SetLevel(Level) error { return nil }

func (discard) GetLevel() Level { return Off }

// MultiLogger is a common logger that can set log level and print log.
type MultiLogger struct {
	writer io.Writer
	level  Level
	rwm    sync.RWMutex
}

// NewMultiLogger is used to create a MultiLogger.
func NewMultiLogger(lv Level, writers ...io.Writer) (Logger, error) {
	lg := MultiLogger{}
	err := lg.SetLevel(lv)
	if err != nil {
		return nil, err
	}
	lg.writer = io.MultiWriter(writers...)
	return &lg, nil
}

// Printf is used to print log with format.
func (ml *MultiLogger) Printf(lv Level, src, format string, log ...interface{}) {
	if ml.discard(lv) {
		return
	}
	buf := Prefix(time.Now(), lv, src)
	_, _ = fmt.Fprintf(buf, format, log...)
	buf.WriteString("\n")
	_, _ = buf.WriteTo(ml.writer)
}

// Print is used to print log.
func (ml *MultiLogger) Print(lv Level, src string, log ...interface{}) {
	if ml.discard(lv) {
		return
	}
	buf := Prefix(time.Now(), lv, src)
	_, _ = fmt.Fprint(buf, log...)
	buf.WriteString("\n")
	_, _ = buf.WriteTo(ml.writer)
}

// Println is used to print log with new line.
func (ml *MultiLogger) Println(lv Level, src string, log ...interface{}) {
	if ml.discard(lv) {
		return
	}
	buf := Prefix(time.Now(), lv, src)
	_, _ = fmt.Fprintln(buf, log...)
	_, _ = buf.WriteTo(ml.writer)
}

func (ml *MultiLogger) discard(lv Level) bool {
	ml.rwm.RLock()
	defer ml.rwm.RUnlock()
	return lv < ml.level
}

// SetLevel is used to set log level that need print.
func (ml *MultiLogger) SetLevel(lv Level) error {
	if lv > Off {
		return fmt.Errorf("invalid logger level: %d", lv)
	}
	ml.rwm.Lock()
	defer ml.rwm.Unlock()
	ml.level = lv
	return nil
}

// GetLevel is used to get the current log level.
func (ml *MultiLogger) GetLevel() Level {
	ml.rwm.RLock()
	defer ml.rwm.RUnlock()
	return ml.level
}
