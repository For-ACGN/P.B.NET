package module

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
)

// Module is the interface of module, it include internal and external module.
//
// Internal module is in the internal/module/*. These module usually use less
// space (use the exist go packages that in GOROOT/src and go.mod), have high
// stability, don't need external program.
//
// External module is in the project/module, or other. These module usually
// have the client(Beacon) and server(external program), client is used to send
// command to the server and receive the result. Client and Server can use Pipe
// or Socket for communication. These module maybe not have high stability and
// will execute high risk operation.
type Module interface {
	// Name is used to get the module name, usually it will not changed,
	// of course, if you want to change it dynamic, don't forget add locker.
	Name() string

	// Description is used to get the module description, usually it will
	// not changed, if you want to change it dynamic, don't forget add locker.
	Description() string

	// Start is used to start module.
	// common module(port scanner): you can initialize some resource.
	// server module(socks5 server): it will start serving.
	// C/S module(mimikatz with RPC): server start serving, client connect it.
	Start() error

	// Stop is used to stop module, it will stop all sub task in module.
	// For example a port scanner module, all scan task will be killed.
	Stop()

	// Restart is used to restart module(usually call inner stop and start).
	Restart() error

	// IsStarted is used to check the module is started.
	IsStarted() bool

	// Info is used to get the module information, usually the return value
	// will not be changed, however it maybe change dynamic. Information will
	// not be displayed default.
	Info() string

	// Status is used to get the module current status, usually the return
	// value will change dynamic. Status will be displayed default.
	Status() string

	// Methods is used to get the definitions of extended methods that can
	// be called by Module.Call function, use String() for get help.
	// The key of the map in return value is the method name, map value is
	// the detailed information include name, parameter and return value.
	Methods() map[string]string

	// Call is used to call extended method, it will return a general value
	// and it maybe include an error(internal/module/plugin) and a general
	// error, usually check the general error firstly, then read the general
	// value and get the error, finally check the inner error.
	Call(method string, arguments ...interface{}) (interface{}, error)
}

// Method contain method function information.
type Method struct {
	Name string   // method name
	Args []*Value // argument
	Rets []*Value // return value
}

// String is used to print method definition.
// output:
// --------------------------------  --------------------------------
// method: Scan                      method: Kill
// --------------------------------  --------------------------------
// parameter:                        return value:
//   host string                       ok  bool
//   port uint16                       err error
// --------------------------------  --------------------------------
// return value:
//   opened bool
//   err    error
// --------------------------------
func (m *Method) String() string {
	buf := bytes.NewBuffer(make([]byte, 0, 64))
	buf.WriteString("--------------------------------\n")
	_, _ = fmt.Fprintf(buf, "method: %s\n", m.Name)
	buf.WriteString("--------------------------------")
	// calculate the max line length about the parameter
	var maxLine int
	for i := 0; i < len(m.Args); i++ {
		l := len(m.Args[i].Name)
		if l > maxLine {
			maxLine = l
		}
	}
	if maxLine != 0 { // has parameters
		buf.WriteString("\n--------------------------------\n")
		buf.WriteString("parameter:\n")
		format := "  %-" + strconv.Itoa(maxLine) + "s %s\n"
		for i := 0; i < len(m.Args); i++ {
			typ := reflect.TypeOf(m.Args[i].Type).String()
			_, _ = fmt.Fprintf(buf, format, m.Args[i].Name, typ)
		}
		buf.WriteString("--------------------------------")
	}
	// calculate the max line length about the return value
	maxLine = 0
	for i := 0; i < len(m.Rets); i++ {
		l := len(m.Rets[i].Name)
		if l > maxLine {
			maxLine = l
		}
	}
	if maxLine != 0 { // has return value
		buf.WriteString("\n--------------------------------\n")
		buf.WriteString("return value:\n")
		format := "  %-" + strconv.Itoa(maxLine) + "s %s\n"
		for i := 0; i < len(m.Rets); i++ {
			typ := reflect.TypeOf(m.Rets[i].Type).String()
			_, _ = fmt.Fprintf(buf, format, m.Rets[i].Name, typ)
		}
		buf.WriteString("--------------------------------")
	}
	return buf.String()
}

// Value is the method argument or return value.
type Value struct {
	Name string      // value name
	Type interface{} // value type
}

// String is used to print value information like "foo string".
func (val *Value) String() string {
	return fmt.Sprintf("%s %s", val.Name, reflect.TypeOf(val.Type).String())
}
