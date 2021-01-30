package convert

import (
	"bytes"
	"io"
	"os"
	"strings"
)

const defaultStringsLineLen = 64

func calcDumpStringWithPLBufferSize(str, prefix string, lineLen int) int {
	sl := len(str)
	pl := len(prefix)
	body := (sl / lineLen) * (pl + lineLen + len(newLine))
	tail := pl + sl%lineLen
	return body + tail
}

// DumpString is used to split string to each line and write it to os.Stdout.
func DumpString(str string) {
	DumpStringWithPL(str, "", defaultStringsLineLen)
}

// SdumpString is used to split string to each line and write it to a string.
func SdumpString(str string) string {
	return SdumpStringWithPL(str, "", defaultStringsLineLen)
}

// FdumpString is used to split string to each line and write it to a io.Writer.
func FdumpString(w io.Writer, str string) (int, error) {
	return FdumpStringWithPL(w, str, "", defaultStringsLineLen)
}

// DumpStringWithPL is used to split string with prefix and line length, write it to os.Stdout.
func DumpStringWithPL(str, prefix string, lineLen int) {
	// calculate buffer size
	size := calcDumpStringWithPLBufferSize(str, prefix, lineLen) + 1
	// dump string
	buf := bytes.NewBuffer(make([]byte, 0, size))
	_, _ = FdumpStringWithPL(buf, str, prefix, lineLen)
	buf.Write(newLine)
	// write to stdout
	_, _ = buf.WriteTo(os.Stdout)
}

// SdumpStringWithPL is used to split string with prefix and line length, write it to a string.
func SdumpStringWithPL(str, prefix string, lineLen int) string {
	// calculate buffer size
	size := calcDumpStringWithPLBufferSize(str, prefix, lineLen)
	// dump string
	builder := strings.Builder{}
	builder.Grow(size)
	_, _ = FdumpStringWithPL(&builder, str, prefix, lineLen)
	// build string
	return builder.String()
}

// FdumpStringWithPL is used to split string with prefix and line length, write it to a io.Writer.
//
// Output:
// ------one line------
// abc12345abc12345abc12345abc12345abc12345abc12345abc12345abc12345
// -------common-------
// abc12345abc12345abc12345abc12345abc12345abc12345abc12345abc12345
// abc12345abc12345abc12345abc12345
// -----with prefix----
//   abc12345abc12345abc12345abc12345abc12345abc12345abc12345abc12345
//   abc12345abc12345abc12345abc12345abc12345abc12345abc12345abc12345
func FdumpStringWithPL(w io.Writer, str, prefix string, lineLen int) (int, error) {
	if len(str) == 0 {
		return 0, nil
	}
	if lineLen < 1 {
		lineLen = defaultStringsLineLen
	}
	hasPrefix := len(prefix) != 0
	var prefixBytes []byte
	if hasPrefix {
		prefixBytes = []byte(prefix)
	}
	reader := strings.NewReader(str)
	buf := make([]byte, lineLen)
	var (
		num int
		nn  int
		n   int
		err error
	)
	for {
		// write prefix
		if hasPrefix {
			nn, err = w.Write(prefixBytes)
			num += nn
			if err != nil {
				return num, err
			}
		}
		// read line and write string
		n, _ = reader.Read(buf)
		nn, err = w.Write(buf[:n])
		num += nn
		if err != nil {
			return num, err
		}
		// finish
		if n != lineLen || reader.Len() == 0 {
			break
		}
		// write new line
		nn, err = w.Write(newLine)
		num += nn
		if err != nil {
			return num, err
		}
	}
	return num, nil
}
