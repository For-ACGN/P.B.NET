package module

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Method contain method function information, it used to generate Methods []string.
type Method struct {
	Name string   // method name
	Desc string   // description
	Args []*Value // argument
	Rets []*Value // return value
}

// Value is the method argument or return value.
type Value struct {
	Name string // value name
	Type string // value type
}

// String is used to print method definition.
// output:
// ------------------------------------------------
// Method: Scan
// ------------------------------------------------
// Description:
//   Scan is used to scan a host with port, it will
//   return the port status.
// ------------------------------------------------
// Parameter:
//   host string
//   port uint16
// ------------------------------------------------
// Return Value:
//   open bool
//   err  error
// ------------------------------------------------
func (m *Method) String() string {
	buf := bytes.NewBuffer(make([]byte, 0, 48*3+64))
	m.printMethodName(buf)
	m.printDescription(buf)
	m.printParameters(buf)
	m.printReturnValue(buf)
	buf.WriteString("------------------------------------------------")
	return buf.String()
}

func (m *Method) printMethodName(buf *bytes.Buffer) {
	buf.WriteString("------------------------------------------------\n")
	_, _ = fmt.Fprintf(buf, "Method: %s\n", m.Name)
	buf.WriteString("------------------------------------------------\n")
}

func (m *Method) printDescription(buf *bytes.Buffer) {
	const lineLen = 48 - 2 // already added "  " before each line
	_, _ = fmt.Fprintln(buf, "Description:")
	if len(m.Desc) == 0 {
		return
	}
	buf.WriteString("  ")
	reader := strings.NewReader(m.Desc)
	line := make([]byte, lineLen)
	for {
		n, _ := io.ReadFull(reader, line)
		buf.Write(line[:n])
		buf.WriteString("\n")
		// has more data
		if n != lineLen {
			break
		}
		// check the next char is " "
		char, err := reader.ReadByte()
		if err != nil { // EOF
			break
		}
		if char == ' ' {
			buf.WriteString(" ")
		} else {
			buf.WriteString("  ")
		}
		buf.WriteByte(char)
	}
}

func (m *Method) printParameters(buf *bytes.Buffer) {
	if len(m.Args) == 0 {
		return
	}
	// calculate the max line length about parameter
	var maxLine int
	for i := 0; i < len(m.Args); i++ {
		l := len(m.Args[i].Name)
		if l > maxLine {
			maxLine = l
		}
	}
	format := "  %-" + strconv.Itoa(maxLine) + "s %s\n"
	// print parameter
	buf.WriteString("------------------------------------------------\n")
	buf.WriteString("Parameter:\n")
	for i := 0; i < len(m.Args); i++ {
		name := m.Args[i].Name
		typ := m.Args[i].Type
		_, _ = fmt.Fprintf(buf, format, name, typ)
	}
}

func (m *Method) printReturnValue(buf *bytes.Buffer) {
	if len(m.Rets) == 0 {
		return
	}
	// calculate the max line length about return value
	var maxLine int
	for i := 0; i < len(m.Rets); i++ {
		l := len(m.Rets[i].Name)
		if l > maxLine {
			maxLine = l
		}
	}
	format := "  %-" + strconv.Itoa(maxLine) + "s %s\n"
	// print return value
	buf.WriteString("------------------------------------------------\n")
	buf.WriteString("Return Value:\n")
	for i := 0; i < len(m.Rets); i++ {
		name := m.Rets[i].Name
		typ := m.Rets[i].Type
		_, _ = fmt.Fprintf(buf, format, name, typ)
	}
}
