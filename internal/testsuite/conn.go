package testsuite

import (
	"crypto/tls"
	"io"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type dialer func() (net.Conn, error)

// Handshaker is used to call connection Handshake(). Some server side connection
// like tls.Conn must call Handshake(), otherwise Dial() will block.
type Handshaker interface {
	Handshake() error
}

// ListenerAndDial is used to test net.Listener and Dial.
func ListenerAndDial(t *testing.T, listener net.Listener, dial dialer, close bool) {
	t.Log("ConnSC")
	for i := 0; i < 3; i++ {
		t.Logf("%d\n", i)
		server, client := AcceptAndDial(t, listener, dial)
		ConnSC(t, server, client, close)
	}
	t.Log("ConnCS")
	for i := 0; i < 3; i++ {
		t.Logf("%d\n", i)
		server, client := AcceptAndDial(t, listener, dial)
		ConnCS(t, client, server, close)
	}
	err := listener.Close()
	require.NoError(t, err)

	IsDestroyed(t, listener)
}

// AcceptAndDial is used to accept and dial a connection.
func AcceptAndDial(t *testing.T, listener net.Listener, dial dialer) (net.Conn, net.Conn) {
	wg := sync.WaitGroup{}
	var server net.Conn
	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		server, err = listener.Accept()
		require.NoError(t, err)
		if s, ok := server.(Handshaker); ok {
			err = s.Handshake()
			require.NoError(t, err)
		}
	}()
	client, err := dial()
	require.NoError(t, err)
	wg.Wait()
	return server, client
}

// if close == true, IsDestroyed will be run after Conn.Close().
// if connection about TLS and use net.Pipe(), set close = false
//
// server, client := net.Pipe()
// tlsServer = tls.Server(server, tlsConfig)
// tlsClient = tls.Client(client, tlsConfig)
// ConnSC(t, tlsServer, tlsClient, false) must set false

// ConnSC is used to test server & client connection,
// server connection will send data firstly.
func ConnSC(t *testing.T, server, client net.Conn, close bool) {
	connAddr(t, server, client)
	conn(t, server, client, close)
}

// ConnCS is used to test client & server connection,
// client connection will send data firstly.
func ConnCS(t *testing.T, client, server net.Conn, close bool) {
	connAddr(t, server, client)
	conn(t, client, server, close)
}

func connAddr(t *testing.T, server, client net.Conn) {
	t.Run("address", func(t *testing.T) {
		t.Log("server remote:", server.RemoteAddr().Network(), server.RemoteAddr())
		t.Log("client local:", client.LocalAddr().Network(), client.LocalAddr())
		t.Log("server local:", server.LocalAddr().Network(), server.LocalAddr())
		t.Log("client remote:", client.RemoteAddr().Network(), client.RemoteAddr())

		// skip udp, because client.LocalAddr() always net.IPv4zero or net.IPv6zero
		if !strings.Contains(server.RemoteAddr().Network(), "udp") {
			require.Equal(t, server.RemoteAddr().Network(), client.LocalAddr().Network())
			require.Equal(t, server.RemoteAddr().String(), client.LocalAddr().String())
		}
		require.Equal(t, server.LocalAddr().Network(), client.RemoteAddr().Network())
		require.Equal(t, server.LocalAddr().String(), client.RemoteAddr().String())
	})
}

// conn1 will send data firstly.
func conn(t *testing.T, conn1, conn2 net.Conn, close bool) {
	// Read(), Write() and SetDeadline()
	write := func(conn net.Conn) {
		data := Bytes()
		n, err := conn.Write(data)
		require.NoError(t, err)
		require.Equal(t, TestDataSize, n)
		require.Equal(t, Bytes(), data)
	}
	read := func(conn net.Conn) {
		data := make([]byte, 256)
		n, err := io.ReadFull(conn, data)
		require.NoError(t, err)
		require.Equal(t, TestDataSize, n)
		require.Equal(t, Bytes(), data)
	}

	wg := sync.WaitGroup{}
	t.Run("read and write", func(t *testing.T) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := conn2.SetDeadline(time.Now().Add(5 * time.Second))
			require.NoError(t, err)
			read(conn2)
			write(conn2)
			wg.Add(2)
			go func() {
				defer wg.Done()
				write(conn2)
			}()
			go func() {
				defer wg.Done()
				write(conn2)
			}()
			read(conn2)
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			err := conn1.SetDeadline(time.Now().Add(5 * time.Second))
			require.NoError(t, err)
			wg.Add(1)
			go func() {
				defer wg.Done()
				write(conn1)
			}()
			read(conn1)
			read(conn1)
			read(conn1)
			wg.Add(1)
			go func() {
				defer wg.Done()
				write(conn1)
			}()
		}()
		wg.Wait()
	})

	// recover about net.Pipe()
	err := conn1.SetDeadline(time.Time{})
	require.NoError(t, err)
	err = conn2.SetDeadline(time.Time{})
	require.NoError(t, err)

	t.Run("read and write parallel", func(t *testing.T) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			read(conn2)
			write(conn2)
			wg.Add(2)
			go func() {
				defer wg.Done()
				write(conn2)
			}()
			go func() {
				defer wg.Done()
				write(conn2)
			}()
			read(conn2)
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			wg.Add(1)
			go func() {
				defer wg.Done()
				write(conn1)
			}()
			read(conn1)
			read(conn1)
			read(conn1)
			wg.Add(1)
			go func() {
				defer wg.Done()
				write(conn1)
			}()
		}()
		wg.Wait()
	})

	t.Run("deadline", func(t *testing.T) {
		err = conn1.SetReadDeadline(time.Now().Add(10 * time.Millisecond))
		require.NoError(t, err)
		err = conn1.SetWriteDeadline(time.Now().Add(10 * time.Millisecond))
		require.NoError(t, err)
		err = conn2.SetReadDeadline(time.Now().Add(10 * time.Millisecond))
		require.NoError(t, err)
		err = conn2.SetWriteDeadline(time.Now().Add(10 * time.Millisecond))
		require.NoError(t, err)
		time.Sleep(30 * time.Millisecond)
		buf := Bytes()
		n, err := conn1.Write(buf)
		require.Error(t, err)
		require.Equal(t, 0, n)
		n, err = conn2.Read(buf)
		require.Error(t, err)
		require.Equal(t, 0, n)

		err = conn1.SetReadDeadline(time.Now().Add(10 * time.Millisecond))
		require.NoError(t, err)
		err = conn1.SetWriteDeadline(time.Now().Add(10 * time.Millisecond))
		require.NoError(t, err)
		err = conn2.SetReadDeadline(time.Now().Add(10 * time.Millisecond))
		require.NoError(t, err)
		err = conn2.SetWriteDeadline(time.Now().Add(10 * time.Millisecond))
		require.NoError(t, err)
		time.Sleep(30 * time.Millisecond)
		buf = Bytes()
		n, err = conn1.Write(buf)
		require.Error(t, err)
		require.Equal(t, 0, n)
		n, err = conn2.Read(buf)
		require.Error(t, err)
		require.Equal(t, 0, n)
	})

	// recover about net.Pipe()
	err = conn1.SetDeadline(time.Time{})
	require.NoError(t, err)
	err = conn2.SetDeadline(time.Time{})
	require.NoError(t, err)

	if !close {
		return
	}
	t.Run("close", func(t *testing.T) {
		buf := Bytes()
		wg.Add(8)
		for i := 0; i < 4; i++ {
			go func() {
				defer wg.Done()
				_, _ = conn1.Write(buf)
			}()
			go func() {
				defer wg.Done()
				_, _ = conn2.Write(buf)
			}()
		}
		// tls.Conn.Close() still send data, so conn2 Close first
		err = conn2.Close()
		require.NoError(t, err)
		err = conn1.Close()
		require.NoError(t, err)
		wg.Wait()

		n, err := conn1.Read(make([]byte, 1024))
		require.Error(t, err)
		require.Equal(t, 0, n)
		n, err = conn2.Read(make([]byte, 1024))
		require.Error(t, err)
		require.Equal(t, 0, n)

		// *tls.Conn start reference self after Go 1.15
		type raw interface { // type: *xnet.Conn
			RawConn() net.Conn
		}
		switch conn1.(type) {
		case *tls.Conn:
			return
		case raw:
			conn := conn1.(raw).RawConn()
			if _, ok := conn.(*tls.Conn); ok {
				return
			}
		}

		IsDestroyed(t, conn1)
		IsDestroyed(t, conn2)
	})
}

// PipeWithReaderWriter is used to call net.Pipe that one side write
// and other side read. reader will block until return.
func PipeWithReaderWriter(t *testing.T, reader, writer func(net.Conn)) {
	server, client := net.Pipe()
	defer func() {
		err := server.Close()
		require.NoError(t, err)
		err = client.Close()
		require.NoError(t, err)
	}()

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		writer(server)
	}()

	reader(client)

	wg.Wait()
}
