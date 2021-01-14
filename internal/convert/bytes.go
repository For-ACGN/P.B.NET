package convert

import (
	"bytes"
	"encoding/hex"
	"io"
	"os"
)

const defaultLineLength = 8

// FdumpBytes is used to convert []byte to go source and dump it to io.Writer.
func FdumpBytes(w io.Writer, b []byte) (int, error) {
	return FdumpBytesWithLineLength(w, b, defaultLineLength)
}

// SdumpBytes is used to convert []byte to go source and dump it to a string.
func SdumpBytes(b []byte) string {
	return SdumpBytesWithLineLength(b, defaultLineLength)
}

// DumpBytes is used to convert []byte to go source and dump it to a os.Stdout.
func DumpBytes(b []byte) {
	DumpBytesWithLineLength(b, defaultLineLength)
}

// FdumpBytesWithLineLength is used to convert []byte to go source with line length and dump it to io.Writer.
func FdumpBytesWithLineLength(w io.Writer, b []byte, l int) (int, error) {
	return w.Write(fdumpBytes(b, l).Bytes())
}

// SdumpBytesWithLineLength is used to convert []byte to go source with line length and dump it to a string.
func SdumpBytesWithLineLength(b []byte, l int) string {
	return fdumpBytes(b, l).String()
}

// DumpBytesWithLineLength is used to convert []byte to go source with line length and dump it to a os.Stdout.
func DumpBytesWithLineLength(b []byte, l int) {
	buf := fdumpBytes(b, l)
	buf.WriteString("\n")
	_, _ = os.Stdout.Write(buf.Bytes())
}

// fdumpBytes is used to convert byte slice to go code, usually it used for go template code.
//
// Output:
// ------one line------
// []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
// -------common-------
// []byte{
//		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
//		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
//		0x00, 0x00, 0x00, 0x00,
// }
// ------full line-----
// []byte{
//		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
//		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
// }
func fdumpBytes(b []byte, lineLength int) *bytes.Buffer {
	const (
		begin = "[]byte{"
		end   = "}"
	)
	// special: empty data
	l := len(b)
	if l == 0 {
		return bytes.NewBuffer([]byte(begin + end))
	}
	// invalid line size
	if lineLength < 1 {
		lineLength = 8
	}
	// create buffer
	bufSize := len(begin+end) + len("0x00, ")*l + l/8
	buf := bytes.NewBuffer(make([]byte, 0, bufSize))
	// write begin string
	buf.WriteString("[]byte{")
	hexBuf := make([]byte, 2)
	// special: one line
	if l <= lineLength {
		for i := 0; i < l; i++ {
			hex.Encode(hexBuf, []byte{b[i]})
			buf.WriteString("0x")
			buf.Write(bytes.ToUpper(hexBuf))
			if i != l-1 {
				buf.WriteString(", ")
			}
		}
		buf.WriteString("}")
		return buf
	}
	// write begin string
	var counter int // need new line
	buf.WriteString("\n")
	for i := 0; i < l; i++ {
		if counter == 0 {
			buf.WriteString("\t")
		}
		hex.Encode(hexBuf, []byte{b[i]})
		buf.WriteString("0x")
		buf.Write(bytes.ToUpper(hexBuf))
		counter++
		if counter == lineLength {
			buf.WriteString(",\n")
			counter = 0
		} else {
			buf.WriteString(", ")
		}
	}
	// write end string
	if counter != 0 { // delete last space
		buf.Truncate(buf.Len() - 1)
		buf.WriteString("\n}")
		return buf
	}
	buf.WriteString("}")
	return buf
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
