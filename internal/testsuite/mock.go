package testsuite

import (
	"bufio"
	"context"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"project/internal/patch/monkey"
)

// errors and panics about mock.
var (
	errMockConnClose = errors.New("mock error in mockConn.Close()")

	errMockListenerAccept = &mockNetError{temporary: true}
	errMockListener       = errors.New("accept more than 10 times")
	errMockListenerClose  = errors.New("mock error in mockListener.Close()")
	mockListenerPanic     = "mock panic in mockListener.Accept()"

	errMockReadCloser = errors.New("mock error in mockReadCloser")
)

// mockNetError implement net.Error.
type mockNetError struct {
	timeout   bool
	temporary bool
}

func (*mockNetError) Error() string {
	return "mock net error"
}

func (e *mockNetError) Timeout() bool {
	return e.timeout
}

func (e *mockNetError) Temporary() bool {
	return e.temporary
}

type mockConnLocalAddr struct{}

func (mockConnLocalAddr) Network() string {
	return "mock Conn local network"
}

func (mockConnLocalAddr) String() string {
	return "mock Conn local address"
}

type mockConnRemoteAddr struct{}

func (mockConnRemoteAddr) Network() string {
	return "mock Conn remote network"
}

func (mockConnRemoteAddr) String() string {
	return "mock Conn remote address"
}

type mockConn struct {
	local  mockConnLocalAddr
	remote mockConnRemoteAddr
	close  bool // close error
}

func (c *mockConn) Read([]byte) (int, error) {
	return 0, nil
}

func (c *mockConn) Write([]byte) (int, error) {
	return 0, nil
}

func (c *mockConn) Close() error {
	if c.close {
		return errMockConnClose
	}
	return nil
}

func (c *mockConn) LocalAddr() net.Addr {
	return c.local
}

func (c *mockConn) RemoteAddr() net.Addr {
	return c.remote
}

func (c *mockConn) SetDeadline(time.Time) error {
	return nil
}

func (c *mockConn) SetReadDeadline(time.Time) error {
	return nil
}

func (c *mockConn) SetWriteDeadline(time.Time) error {
	return nil
}

// NewMockConnWithCloseError is used to create a mock conn
// that will return a errMockConnClose when call Close().
func NewMockConnWithCloseError() net.Conn {
	return &mockConn{close: true}
}

// IsMockConnCloseError is used to check err is errMockConnClose.
func IsMockConnCloseError(t testing.TB, err error) {
	require.Equal(t, errMockConnClose, err)
}

type mockListenerAddr struct{}

func (mockListenerAddr) Network() string {
	return "mock Listener network"
}

func (mockListenerAddr) String() string {
	return "mock Listener address"
}

type mockListener struct {
	addr  mockListenerAddr
	error bool // accept error
	panic bool // accept panic
	close bool // close error
	n     int  // accept count
}

func (l *mockListener) Accept() (net.Conn, error) {
	if l.n > 10 {
		return nil, errMockListener
	}
	l.n++
	if l.error {
		return nil, errMockListenerAccept
	}
	if l.panic {
		panic(mockListenerPanic)
	}
	return nil, nil
}

func (l *mockListener) Close() error {
	if l.close {
		return errMockListenerClose
	}
	return nil
}

func (l *mockListener) Addr() net.Addr {
	return l.addr
}

// NewMockListenerWithError is used to create a mock listener
// that return a custom error call Accept().
func NewMockListenerWithError() net.Listener {
	return &mockListener{error: true}
}

// NewMockListenerWithPanic is used to create a mock listener
// that panic when call Accept().
func NewMockListenerWithPanic() net.Listener {
	return &mockListener{panic: true}
}

// NewMockListenerWithCloseError is used to create a mock listener
// that will return a errMockListenerClose when call Close().
func NewMockListenerWithCloseError() net.Listener {
	return &mockListener{close: true}
}

// IsMockListenerError is used to check err is errMockListenerAccept.
func IsMockListenerError(t testing.TB, err error) {
	require.Equal(t, errMockListener, err)
}

// IsMockListenerPanic is used to check err.Error() is mockListenerPanic.
func IsMockListenerPanic(t testing.TB, err error) {
	require.Contains(t, err.Error(), mockListenerPanic)
}

// IsMockListenerCloseError is used to check err is errMockListenerClose.
func IsMockListenerCloseError(t testing.TB, err error) {
	require.Equal(t, errMockListenerClose, err)
}

type mockResponseWriter struct {
	hijack bool
	conn   net.Conn
}

func (mockResponseWriter) Header() http.Header {
	return nil
}

func (mockResponseWriter) Write([]byte) (int, error) {
	return 0, nil
}

func (mockResponseWriter) WriteHeader(int) {}

func (rw mockResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if rw.hijack {
		return nil, nil, errors.New("failed to hijack")
	}
	return rw.conn, nil, nil
}

// NewMockResponseWriter is used to create simple mock response writer.
func NewMockResponseWriter() http.ResponseWriter {
	server, client := net.Pipe()
	go func() { _, _ = io.Copy(ioutil.Discard, server) }()
	return &mockResponseWriter{conn: client}
}

// NewMockResponseWriterWithFailedToHijack is used to create a mock
// http.ResponseWriter that implemented http.Hijacker, if call Hijack()
// it will return an error.
func NewMockResponseWriterWithFailedToHijack() http.ResponseWriter {
	return &mockResponseWriter{hijack: true}
}

// NewMockResponseWriterWithFailedToWrite is used to create a mock
// http.ResponseWriter that implemented http.Hijacker, if use hijacked
// connection, it will return an error.
func NewMockResponseWriterWithFailedToWrite() http.ResponseWriter {
	server, client := net.Pipe()
	_ = client.Close()
	_ = server.Close()
	return &mockResponseWriter{conn: client}
}

type mockConnClosePanic struct {
	net.Conn
	server net.Conn
}

func (c *mockConnClosePanic) Close() error {
	defer func() { panic("mock panic in Close()") }()
	_ = c.Conn.Close()
	_ = c.server.Close()
	return nil
}

// NewMockResponseWriterWithClosePanic is used to create a mock
// http.ResponseWriter that implemented http.Hijacker, if use hijacked
// connection and when call Close() it will panic.
func NewMockResponseWriterWithClosePanic() http.ResponseWriter {
	server, client := net.Pipe()
	go func() { _, _ = io.Copy(ioutil.Discard, server) }()
	mc := mockConnClosePanic{
		Conn:   client,
		server: server,
	}
	return &mockResponseWriter{conn: &mc}
}

type mockConnReadPanic struct {
	net.Conn
	server net.Conn
}

func (c *mockConnReadPanic) Read([]byte) (int, error) {
	panic("mock panic in Read()")
}

func (c *mockConnReadPanic) Close() error {
	_ = c.Conn.Close()
	_ = c.server.Close()
	return nil
}

// DialMockConnWithReadPanic is used to create a mock connection
// and when call Read() it will panic.
func DialMockConnWithReadPanic(_ context.Context, _, _ string) (net.Conn, error) {
	server, client := net.Pipe()
	go func() { _, _ = io.Copy(ioutil.Discard, server) }()
	return &mockConnReadPanic{
		Conn:   client,
		server: server,
	}, nil
}

type mockConnWriteError struct {
	net.Conn
	server net.Conn
}

func (c *mockConnWriteError) Read(b []byte) (int, error) {
	b[0] = 1
	return 1, nil
}

func (c *mockConnWriteError) Write([]byte) (int, error) {
	return 0, monkey.Error
}

func (c *mockConnWriteError) Close() error {
	_ = c.Conn.Close()
	_ = c.server.Close()
	return nil
}

// DialMockConnWithWriteError is used to create a mock connection
// and when call Write() it will return a monkey error.
func DialMockConnWithWriteError(_ context.Context, _, _ string) (net.Conn, error) {
	server, client := net.Pipe()
	go func() { _, _ = io.Copy(ioutil.Discard, server) }()
	return &mockConnWriteError{
		Conn:   client,
		server: server,
	}, nil
}

type mockReadCloser struct {
	panic bool
	rwm   sync.RWMutex
}

func (rc *mockReadCloser) Read([]byte) (int, error) {
	rc.rwm.RLock()
	defer rc.rwm.RUnlock()
	if rc.panic {
		panic("mock panic in Read()")
	}
	return 0, errMockReadCloser
}

func (rc *mockReadCloser) Close() error {
	rc.rwm.Lock()
	defer rc.rwm.Unlock()
	rc.panic = false
	return nil
}

// NewMockReadCloserWithReadError is used to return a ReadCloser that
// return a errMockReadCloser when call Read().
func NewMockReadCloserWithReadError() io.ReadCloser {
	return new(mockReadCloser)
}

// NewMockReadCloserWithReadPanic is used to return a ReadCloser that
// panic when call Read().
func NewMockReadCloserWithReadPanic() io.ReadCloser {
	return &mockReadCloser{panic: true}
}

// IsMockReadCloserError is used to confirm err is errMockReadCloser.
func IsMockReadCloserError(t testing.TB, err error) {
	require.Equal(t, errMockReadCloser, err)
}
