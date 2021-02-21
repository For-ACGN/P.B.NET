package logger

import (
	"bytes"
	"fmt"
	"strings"
	"time"
)

// Level is the log level.
type Level uint8

func (lv Level) String() string {
	switch lv {
	case Trace:
		return "trace"
	case Debug:
		return "debug"
	case Info:
		return "info"
	case Critical:
		return "critical"
	case Warning:
		return "warning"
	case Error:
		return "error"
	case Exploit:
		return "exploit"
	case Fatal:
		return "fatal"
	default:
		return "unknown"
	}
}

// about log levels.
const (
	All Level = iota // log all(include unexpected level)

	Trace // for trace function (development)
	Debug // generic debug information

	Info     // generic running information
	Critical // important information like exploit successfully

	Warning // appear error but can continue
	Error   // appear error that can not continue (returned)
	Exploit // appear excepted exploit or security problem(maybe)
	Fatal   // appear panic in goroutine or error that need interrupt

	Off // stop logger
)

// ParseLevel is used to parse log level from string.
func ParseLevel(level string) (Level, error) {
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
	case "critical":
		lv = Critical
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

// DumpPrefix is used to print time, level and source to a *bytes.buffer.
// The output is [time] + [level] + <source> + log1 + log2 ...
// Log source is: class name + ":" + instance tag like "server:tag1",
// if log source include more that two words, connect these with "-"
// like "socks5-server:tag1".
//
// [2018-11-27 09:16:16 +08:00] [info] <main> controller is running
func DumpPrefix(t time.Time, lv Level, src string) *bytes.Buffer {
	buf := bytes.Buffer{}
	buf.WriteString("[")
	buf.WriteString(t.Format(TimeLayout))
	buf.WriteString("] [")
	buf.WriteString(lv.String())
	buf.WriteString("] <")
	buf.WriteString(src)
	buf.WriteString("> ")
	return &buf
}
