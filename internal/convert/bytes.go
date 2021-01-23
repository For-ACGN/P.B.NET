package convert

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
)

const defaultBytesLineLen = 8

// DumpBytes is used to convert byte slice to go source and write it to os.Stdout.
func DumpBytes(b []byte) (int, error) {
	n1, _ := fmt.Println("[]byte{")
	n2, err := DumpBytesWithPL(b, "\t", defaultBytesLineLen)
	n3, _ := fmt.Println("\n}")
	return n1 + n2 + n3, err
}

// SdumpBytes is used to convert byte slice to go source and write it to a string.
func SdumpBytes(b []byte) string {
	return
}

// FdumpBytes is used to convert byte slice to go source and write it to io.Writer.
func FdumpBytes(w io.Writer, b []byte) (int, error) {
	return
}

// DumpBytesWithPL is used to convert byte slice to go source code with prefix
// and line length, then write it to os.Stdout.
func DumpBytesWithPL(b []byte, prefix string, lineLen int) (int, error) {
	// calculate buffer size
	bl := len(b)
	pl := len(prefix)
	body := (bl / lineLen) * (pl + lineLen*6 - 1 + len(newLine))
	tail := pl + (bl%lineLen)*6
	// dump string
	buf := bytes.NewBuffer(make([]byte, 0, body+tail+1))
	_, _ = FdumpBytesWithPL(buf, b, prefix, lineLen)
	buf.Write(newLine)
	// write to stdout
	n, err := buf.WriteTo(os.Stdout)
	return int(n), err
}

// SdumpBytesWithPL is used to convert byte slice to go source code with prefix
// and line length, then write it to a string.
func SdumpBytesWithPL(b []byte, prefix string, lineLen int) string {
	// calculate buffer size
	bl := len(b)
	pl := len(prefix)
	body := (bl / lineLen) * (pl + lineLen*6 - 1 + len(newLine))
	tail := pl + (bl%lineLen)*6
	// dump string
	builder := strings.Builder{}
	builder.Grow(body + tail)
	_, _ = FdumpBytesWithPL(&builder, b, prefix, lineLen)
	return builder.String()
}

var newLine = []byte("\n")

// FdumpBytesWithPL is used to convert byte slice to go source code with prefix
// and line length, then write it to io.Writer.
//
// Output:
// ------one line------
// 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
// -------common-------
// 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
// 0x00, 0x00, 0x00, 0x00,
// -----with prefix----
//     0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
//     0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
func FdumpBytesWithPL(w io.Writer, b []byte, prefix string, lineLen int) (int, error) {
	l := len(b)
	if l == 0 {
		return 0, nil
	}
	if lineLen < 1 {
		lineLen = defaultBytesLineLen
	}
	hasPrefix := len(prefix) != 0
	var prefixBytes []byte
	if hasPrefix {
		prefixBytes = []byte(prefix)
	}
	reader := bytes.NewReader(b)
	buf := make([]byte, lineLen)
	hexBuf := make([]byte, 2)
	byt := []byte("0xFF, ")
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
		// read line
		n, _ = reader.Read(buf)
		for i := 0; i < n; i++ {
			hex.Encode(hexBuf, []byte{buf[i]})
			hexBuf = bytes.ToUpper(hexBuf)
			copy(byt[2:], hexBuf)
			// need last space
			if i == lineLen-1 || ((n != lineLen) && (i == n-1)) {
				nn, err = w.Write(byt[:5])
			} else {
				nn, err = w.Write(byt)
			}
			num += nn
			if err != nil {
				return num, err
			}
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

// MergeBytes is used to merge multi bytes slice to one, it will deep copy each slice.
func MergeBytes(bs ...[]byte) []byte {
	n := len(bs)
	if n == 0 {
		return nil
	}
	var l int
	for i := 0; i < n; i++ {
		l += len(bs[i])
	}
	b := make([]byte, 0, l)
	for i := 0; i < n; i++ {
		b = append(b, bs[i]...)
	}
	return b
}
