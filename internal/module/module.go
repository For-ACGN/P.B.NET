package module

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
	// The detailed information include name, parameter and return value.
	Methods() []string

	// Call is used to call extended method, it will return a general value
	// and it maybe include an error(return true, nil) and a general error
	// (like call method is not exist), usually check the general error firstly,
	// then read the general value and get the error, finally check the error.
	Call(method string, args ...interface{}) (interface{}, error)
}
