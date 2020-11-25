package module

// Module is the interface of module, it include internal and external module.
//
// Internal module is in the internal/module/*. These module usually use less
// space (use the exist go packages that in GOROOT/src and go.mod), have high
// stability, don't need external program.
//
// External module is in the project/module, or app/mod. These module usually
// have the client(Beacon) and server(external program), client is used to send
// command to the server and receive the result. Client and Server can use Pipe
// or Socket for communication. These module maybe not have high stability and
// execute high risk operation.
// Use Start() to connect the module server, and use Call() to send command.
type Module interface {
	Start() error
	Stop()
	Restart() error
	Name() string
	Info() string
	Status() string

	// Stopped() bool
}
