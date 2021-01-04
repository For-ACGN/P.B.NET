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
	All Level = iota // show all log messages

	// about debug
	Trace // for trace function (development)
	Debug // general debug information

	// about information
	Info     // common running information
	Critical // important information like attack successfully

	// about error
	Warning // appear error but can continue
	Error   // appear error that can not continue (returned)
	Exploit // find attack exploit or security problem(maybe)
	Fatal   // appear panic in goroutine

	Off // stop log message
)

// TimeLayout is used to provide a parameter to time.Time.Format().
const TimeLayout = "2006-01-02 15:04:05 Z07:00"

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

// Prefix is used to print time, level and source to a buffer.
//
// time + level + source + log
// source usually like: class name + "-" + instance tag
//
// [2018-11-27 00:00:00 +08:00] [info] <main> controller is running
func Prefix(time time.Time, level Level, src string) *bytes.Buffer {
	buf := bytes.Buffer{}
	buf.WriteString("[")
	buf.WriteString(time.Local().Format(TimeLayout))
	buf.WriteString("] [")
	buf.WriteString(level.String())
	buf.WriteString("] <")
	buf.WriteString(src)
	buf.WriteString("> ")
	return &buf
}
