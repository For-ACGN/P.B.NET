package logger

import (
	"bytes"
	"fmt"
	"io"
	"sync"
	"time"
)

// TimeLayout is the parameter about function time.Time.Format().
const TimeLayout = "2006-01-02 15:04:05 Z07:00"

// Logger is used to print log with level and source. If log
// level is lower than current level, this log will be discard.
type Logger interface {
	// Printf is used to log with format, reference features about fmt.Printf.
	Printf(lv Level, src, format string, log ...interface{})

	// Print is used to log with new line, reference features about fmt.Print.
	Print(lv Level, src string, log ...interface{})

	// Println is used to log with new line, reference features about fmt.Println.
	Println(lv Level, src string, log ...interface{})

	// GetLevel is used to get logger current minimum log level.
	GetLevel() Level

	// SetLevel is used to set logger minimum log level.
	SetLevel(lv Level) error
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
	lv  Level
	rwm sync.RWMutex
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
	output := DumpPrefix(time.Now(), lv, src)
	_, _ = fmt.Fprintf(output, format, log...)
	fmt.Println(output)
}

func (c *common) Print(lv Level, src string, log ...interface{}) {
	if c.discard(lv) {
		return
	}
	output := DumpPrefix(time.Now(), lv, src)
	_, _ = fmt.Fprint(output, log...)
	fmt.Println(output)
}

func (c *common) Println(lv Level, src string, log ...interface{}) {
	if c.discard(lv) {
		return
	}
	output := DumpPrefix(time.Now(), lv, src)
	_, _ = fmt.Fprintln(output, log...)
	fmt.Print(output)
}

func (c *common) discard(lv Level) bool {
	c.rwm.RLock()
	defer c.rwm.RUnlock()
	return lv < c.lv
}

func (c *common) GetLevel() Level {
	c.rwm.RLock()
	defer c.rwm.RUnlock()
	return c.lv
}

func (c *common) SetLevel(lv Level) error {
	if lv > Off {
		return fmt.Errorf("invalid logger level: %d", lv)
	}
	c.rwm.Lock()
	defer c.rwm.Unlock()
	c.lv = lv
	return nil
}

// [Test] [2020-01-21 12:36:41 +08:00] [debug] <test src> test-format test log
type test struct {
	lv  Level
	rwm sync.RWMutex
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
	return lv < t.lv
}

func writePrefix(lv Level, src string) *bytes.Buffer {
	output := new(bytes.Buffer)
	output.WriteString("[Test] ")
	_, _ = DumpPrefix(time.Now(), lv, src).WriteTo(output)
	return output
}

func (t *test) GetLevel() Level {
	t.rwm.RLock()
	defer t.rwm.RUnlock()
	return t.lv
}

func (t *test) SetLevel(lv Level) error {
	if lv > Off {
		return fmt.Errorf("invalid logger level: %d", lv)
	}
	t.rwm.Lock()
	defer t.rwm.Unlock()
	t.lv = lv
	return nil
}

type discard struct{}

func (discard) Printf(Level, string, string, ...interface{}) {}

func (discard) Print(Level, string, ...interface{}) {}

func (discard) Println(Level, string, ...interface{}) {}

func (discard) GetLevel() Level { return Off }

func (discard) SetLevel(Level) error { return nil }

var _ Logger = new(multiLogger)

// multiLogger is used to print log to multi io.Writer.
type multiLogger struct {
	writer io.Writer
	lv     Level
	rwm    sync.RWMutex
}

// NewMultiLogger is used to create a multi logger.
func NewMultiLogger(lv Level, writers ...io.Writer) (Logger, error) {
	lg := multiLogger{}
	err := lg.SetLevel(lv)
	if err != nil {
		return nil, err
	}
	lg.writer = io.MultiWriter(writers...)
	return &lg, nil
}

func (ml *multiLogger) Printf(lv Level, src, format string, log ...interface{}) {
	if ml.discard(lv) {
		return
	}
	buf := DumpPrefix(time.Now(), lv, src)
	_, _ = fmt.Fprintf(buf, format, log...)
	buf.WriteString("\n")
	_, _ = buf.WriteTo(ml.writer)
}

func (ml *multiLogger) Print(lv Level, src string, log ...interface{}) {
	if ml.discard(lv) {
		return
	}
	buf := DumpPrefix(time.Now(), lv, src)
	_, _ = fmt.Fprint(buf, log...)
	buf.WriteString("\n")
	_, _ = buf.WriteTo(ml.writer)
}

func (ml *multiLogger) Println(lv Level, src string, log ...interface{}) {
	if ml.discard(lv) {
		return
	}
	buf := DumpPrefix(time.Now(), lv, src)
	_, _ = fmt.Fprintln(buf, log...)
	_, _ = buf.WriteTo(ml.writer)
}

func (ml *multiLogger) discard(lv Level) bool {
	ml.rwm.RLock()
	defer ml.rwm.RUnlock()
	return lv < ml.lv
}

func (ml *multiLogger) GetLevel() Level {
	ml.rwm.RLock()
	defer ml.rwm.RUnlock()
	return ml.lv
}

func (ml *multiLogger) SetLevel(lv Level) error {
	if lv > Off {
		return fmt.Errorf("invalid logger level: %d", lv)
	}
	ml.rwm.Lock()
	defer ml.rwm.Unlock()
	ml.lv = lv
	return nil
}
