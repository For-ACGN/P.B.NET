package convert

import (
	"bytes"
	"encoding/hex"
	"io"
	"os"
)

const defaultLineSize = 8

// FdumpBytes is used to convert []byte to go source and dump it to io.Writer.
func FdumpBytes(w io.Writer, b []byte) {
	FdumpBytesWithLineSize(w, b, defaultLineSize)
}

// SdumpBytes is used to convert []byte to go source and dump it to a string.
func SdumpBytes(b []byte) string {
	return SdumpBytesWithLineSize(b, defaultLineSize)
}

// DumpBytes is used to convert []byte to go source and dump it to a os.Stdout.
func DumpBytes(b []byte) {
	DumpBytesWithLineSize(b, defaultLineSize)
}

// FdumpBytesWithLineSize is used to convert []byte to go source and dump it to io.Writer.
func FdumpBytesWithLineSize(w io.Writer, b []byte, lineSize int) {
	fdumpBytes(w, b, lineSize)
}

// SdumpBytesWithLineSize is used to convert []byte to go source and dump it to a string.
func SdumpBytesWithLineSize(b []byte, lineSize int) string {
	buf := bytes.NewBuffer(make([]byte, 0, (6+1)*len(b)))
	fdumpBytes(buf, b, lineSize)
	return buf.String()
}

// DumpBytesWithLineSize is used to convert []byte to go source and dump it to a os.Stdout.
func DumpBytesWithLineSize(b []byte, lineSize int) {
	fdumpBytes(os.Stdout, b, lineSize)
}

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
func fdumpBytes(w io.Writer, b []byte, lineSize int) {
	const (
		begin = "[]byte{"
		end   = "}"
	)
	// special: empty data
	l := len(b)
	if l == 0 {
		_, _ = w.Write([]byte(begin + end))
		return
	}
	// invalid line size
	if lineSize < 1 {
		lineSize = 8
	}
	// create buffer
	bufSize := len(begin+end) + len("0x00, ")*l + l/8
	buf := bytes.NewBuffer(make([]byte, 0, bufSize))
	// write begin string
	buf.WriteString("[]byte{")
	hexBuf := make([]byte, 2)
	// special: one line
	if l <= lineSize {
		for i := 0; i < l; i++ {
			hex.Encode(hexBuf, []byte{b[i]})
			buf.WriteString("0x")
			buf.Write(bytes.ToUpper(hexBuf))
			if i != l-1 {
				buf.WriteString(", ")
			}
		}
		buf.WriteString("}")
		_, _ = buf.WriteTo(w)
		return
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
		if counter == lineSize {
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
		_, _ = buf.WriteTo(w)
		return
	}
	buf.WriteString("}")
	_, _ = buf.WriteTo(w)
}
