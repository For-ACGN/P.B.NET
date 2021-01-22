package convert

import (
	"bytes"
	"io"
	"os"
	"strings"
)

const defaultStringsLineLen = 64

// FdumpStringWithPL is used to dump string each line to io.Writer.
func FdumpString(w io.Writer, str string) (int, error) {
	return FdumpStringWithPL(w, str, "", defaultStringsLineLen)
}

// SdumpString is used to dump string each line to a string.
func SdumpString(str string) string {
	return SdumpStringWithPL(str, "", defaultStringsLineLen)
}

// DumpString is used to dump string each line to os.Stdout.
func DumpString(str string) (int, error) {
	return DumpStringWithPL(str, "", defaultStringsLineLen)
}

// FdumpStringWithPL is used to dump string each line with prefix and line length to io.Writer.
func FdumpStringWithPL(w io.Writer, str, prefix string, lineLen int) (int, error) {
	return fdumpString(w, str, prefix, lineLen)
}

// FdumpStringWithPL is used to dump string each line with prefix and line length to a string.
func SdumpStringWithPL(str, prefix string, lineLen int) string {
	// calculate buffer size
	sl := len(str)
	pl := len(prefix)
	body := (sl / lineLen) * (pl + lineLen + len(newLine))
	tail := pl + sl%lineLen
	builder := strings.Builder{}
	// dump string
	builder.Grow(body + tail)
	_, _ = fdumpString(&builder, str, prefix, lineLen)
	return builder.String()
}

// DumpStringWithPL is used to dump string each line with prefix and line length to os.Stdout.
func DumpStringWithPL(str, prefix string, lineLen int) (int, error) {
	// calculate buffer size
	sl := len(str)
	pl := len(prefix)
	body := (sl / lineLen) * (pl + lineLen + len(newLine))
	tail := pl + sl%lineLen
	buf := bytes.NewBuffer(make([]byte, 0, body+tail+1))
	// dump string
	_, _ = fdumpString(buf, str, prefix, lineLen)
	buf.Write(newLine)
	n, err := buf.WriteTo(os.Stdout)
	return int(n), err
}

var newLine = []byte("\n")

// fdumpString is used to dump string each line with prefix.
//
// Output:
// ------one line------
// abc12345abc12345abc12345abc12345abc12345abc12345abc12345abc12345
// -------common-------
// abc12345abc12345abc12345abc12345abc12345abc12345abc12345abc12345
// abc12345abc12345abc12345abc12345
// ------with prefix-----
//   abc12345abc12345abc12345abc12345abc12345abc12345abc12345abc12345
//   abc12345abc12345abc12345abc12345abc12345abc12345abc12345abc12345
func fdumpString(w io.Writer, str, prefix string, lineLen int) (int, error) {
	if len(str) == 0 {
		return 0, nil
	}
	hasPrefix := len(prefix) != 0
	var prefixBytes []byte
	if hasPrefix {
		prefixBytes = []byte(prefix)
	}
	var num int
	reader := strings.NewReader(str)
	buf := make([]byte, lineLen)
	for {
		// write prefix
		if hasPrefix {
			nn, err := w.Write(prefixBytes)
			num += nn
			if err != nil {
				return num, err
			}
		}
		// read line
		n, _ := reader.Read(buf)
		nn, err := w.Write(buf[:n])
		num += nn
		if err != nil {
			return num, err
		}
		// finish
		if n != lineLen {
			break
		}
		// write new line
		if reader.Len() != 0 {
			nn, err = w.Write(newLine)
			num += nn
			if err != nil {
				return num, err
			}
		} else {
			break
		}
	}
	return num, nil
}
