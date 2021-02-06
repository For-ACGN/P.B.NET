package convert

import (
	"bytes"
	"encoding/hex"
	"io"
	"os"
	"strings"
)

const defaultBytesLineLen = 8

var newLine = []byte("\n")

func calcDumpBytesWithPLBufferSize(b []byte, prefix string, lineLen int) int {
	bl := len(b)
	pl := len(prefix)
	body := (bl / lineLen) * (pl + lineLen*len("0xFF, ") - 1 + len(newLine))
	tail := pl + (bl%lineLen)*len("0xFF, ")
	return body + tail
}

func calcDumpBytesBufferSize(b []byte) int {
	var prefix string
	needNewLine := len(b) > defaultBytesLineLen
	if needNewLine {
		prefix = "\t"
	}
	size := calcDumpBytesWithPLBufferSize(b, prefix, defaultBytesLineLen) +
		len("[]byte{") + len("}")
	if needNewLine {
		size += 2 * len(newLine)
	}
	return size
}

// DumpBytes is used to convert byte slice to go source and write it to os.Stdout.
func DumpBytes(b []byte) {
	// calculate buffer size
	size := calcDumpBytesBufferSize(b)
	// dump bytes
	buf := bytes.NewBuffer(make([]byte, 0, size))
	_, _ = FdumpBytes(buf, b)
	buf.Write(newLine)
	// write to stdout
	_, _ = buf.WriteTo(os.Stdout)
}

// SdumpBytes is used to convert byte slice to go source and write it to a string.
func SdumpBytes(b []byte) string {
	// calculate buffer size
	size := calcDumpBytesBufferSize(b)
	// dump bytes
	builder := strings.Builder{}
	builder.Grow(size)
	_, _ = FdumpBytes(&builder, b)
	// build string
	return builder.String()
}

// FdumpBytes is used to convert byte slice to go source and write it to a io.Writer.
func FdumpBytes(w io.Writer, b []byte) (int, error) {
	var (
		num int
		n   int
		err error
	)
	// write begin
	n, err = w.Write([]byte("[]byte{"))
	num += n
	if err != nil {
		return num, err
	}
	var prefix string
	needNewLine := len(b) > defaultBytesLineLen
	if needNewLine {
		prefix = "\t"
		n, err = w.Write(newLine)
		num += n
		if err != nil {
			return num, err
		}
	}
	// write body
	n, err = FdumpBytesWithPL(w, b, prefix, defaultBytesLineLen)
	num += n
	if err != nil {
		return num, err
	}
	if needNewLine {
		n, err = w.Write(newLine)
		num += n
		if err != nil {
			return num, err
		}
	}
	// write end
	n, err = w.Write([]byte("}"))
	num += n
	return num, err
}

// DumpBytesWithPL is used to convert byte slice to go source code with prefix
// and line length, then write it to os.Stdout.
func DumpBytesWithPL(b []byte, prefix string, lineLen int) {
	// calculate buffer size
	size := calcDumpBytesWithPLBufferSize(b, prefix, lineLen) + 1
	// dump string
	buf := bytes.NewBuffer(make([]byte, 0, size))
	_, _ = FdumpBytesWithPL(buf, b, prefix, lineLen)
	buf.Write(newLine)
	// write to stdout
	_, _ = buf.WriteTo(os.Stdout)
}

// SdumpBytesWithPL is used to convert byte slice to go source code with prefix
// and line length, then write it to a string.
func SdumpBytesWithPL(b []byte, prefix string, lineLen int) string {
	// calculate buffer size
	size := calcDumpBytesWithPLBufferSize(b, prefix, lineLen)
	// dump string
	builder := strings.Builder{}
	builder.Grow(size)
	_, _ = FdumpBytesWithPL(&builder, b, prefix, lineLen)
	// build string
	return builder.String()
}

// FdumpBytesWithPL is used to convert byte slice to go source code with prefix
// and line length, then write it to a io.Writer.
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
		// write each byte
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
