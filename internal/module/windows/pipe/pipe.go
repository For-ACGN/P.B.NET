// +build windows

package pipe

import (
	"context"
	"net"
	"time"

	"github.com/Microsoft/go-winio"
)

// ErrPipeListenerClosed is returned for pipe operations on listeners that have been closed.
// This error should match net.errClosing since docker takes a dependency on its text.
var ErrPipeListenerClosed = winio.ErrPipeListenerClosed

// Config contain configuration for the pipe listener.
type Config struct {
	// SecurityDescriptor contains a Windows security descriptor in SDDL format.
	SecurityDescriptor string `toml:"security_descriptor"`

	// MessageMode determines whether the pipe is in byte or message mode. In either
	// case the pipe is read in byte mode by default. The only practical difference in
	// this implementation is that CloseWrite() is only supported for message mode pipes;
	// CloseWrite() is implemented as a zero-byte write, but zero-byte writes are only
	// transferred to the reader (and returned as io.EOF in this implementation)
	// when the pipe is in message mode.
	MessageMode bool `toml:"message_mode"`

	// InputBufferSize specifies the size of the input buffer, in bytes.
	InputBufferSize int32 `toml:"input_buffer_size"`

	// OutputBufferSize specifies the size of the output buffer, in bytes.
	OutputBufferSize int32 `toml:"output_buffer_size"`
}

// ListenPipe creates a listener on a Windows named pipe path, e.g. \\.\pipe\test.
// The pipe must not already exist.
func ListenPipe(path string, cfg *Config) (net.Listener, error) {
	if cfg == nil {
		return winio.ListenPipe(path, nil)
	}
	pipeCfg := &winio.PipeConfig{
		SecurityDescriptor: cfg.SecurityDescriptor,
		MessageMode:        cfg.MessageMode,
		InputBufferSize:    cfg.InputBufferSize,
		OutputBufferSize:   cfg.OutputBufferSize,
	}

	return winio.ListenPipe(path, pipeCfg)
}

// DialPipe connects to a named pipe by path, timing out if the connection
// takes longer than the specified duration. If timeout is nil, then we use
// a default timeout of 2 seconds. (We do not use WaitNamedPipe.)
func DialPipe(path string, timeout *time.Duration) (net.Conn, error) {
	return winio.DialPipe(path, timeout)
}

// DialPipeContext attempts to connect to a named pipe by `path`
// until `ctx` cancellation or timeout.
func DialPipeContext(ctx context.Context, path string) (net.Conn, error) {
	return winio.DialPipeContext(ctx, path)
}

// DialPipeAccess attempts to connect to a named pipe by `path` with
// `access` until `ctx` cancellation or timeout.
func DialPipeAccess(ctx context.Context, path string, access uint32) (net.Conn, error) {
	return winio.DialPipeAccess(ctx, path, access)
}
