package testsuite

import (
	"io"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// ListenerAndDial is used to test net.Listener and Dial
func ListenerAndDial(t testing.TB, l net.Listener, d func() (net.Conn, error), close bool) {
	wg := sync.WaitGroup{}
	for i := 0; i < 3; i++ {
		var server net.Conn
		wg.Add(1)
		go func() {
			defer wg.Done()
			var err error
			server, err = l.Accept()
			require.NoError(t, err)
		}()
		client, err := d()
		require.NoError(t, err)
		wg.Wait()
		ConnSC(t, server, client, close)

		t.Log("") // new line for Conn
	}
	for i := 0; i < 3; i++ {
		var server net.Conn
		wg.Add(1)
		go func() {
			defer wg.Done()
			var err error
			server, err = l.Accept()
			require.NoError(t, err)
		}()
		client, err := d()
		require.NoError(t, err)
		wg.Wait()
		ConnCS(t, client, server, close)

		t.Log("") // new line for Conn
	}
	require.NoError(t, l.Close())
	IsDestroyed(t, l)
}

// if close == true, IsDestroyed will be run after Conn.Close()
// if connection about TLS and use net.Pipe(), set close = false
//
// server, client := net.Pipe()
// tlsServer = tls.Server(server, tlsConfig)
// tlsClient = tls.Client(client, tlsConfig)
// ConnSC(t, tlsServer, tlsClient, false) must set false

// ConnSC is used to test server & client connection,
// server connection will send data firstly
func ConnSC(t testing.TB, server, client net.Conn, close bool) {
	connAddr(t, server, client)
	conn(t, server, client, close)
}

// ConnCS is used to test client & server connection,
// client connection will send data firstly
func ConnCS(t testing.TB, client, server net.Conn, close bool) {
	connAddr(t, server, client)
	conn(t, client, server, close)
}

func connAddr(t testing.TB, server, client net.Conn) {
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
}

// conn1 will send data firstly
func conn(t testing.TB, conn1, conn2 net.Conn, close bool) {
	// Read(), Write() and SetDeadline()
	write := func(conn net.Conn) {
		data := Bytes()
		_, err := conn.Write(data)
		require.NoError(t, err)
		require.Equal(t, Bytes(), data)
	}
	read := func(conn net.Conn) {
		data := make([]byte, 256)
		_, err := io.ReadFull(conn, data)
		require.NoError(t, err)
		require.Equal(t, Bytes(), data)
	}
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		require.NoError(t, conn2.SetDeadline(time.Now().Add(5*time.Second)))
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
		require.NoError(t, conn1.SetDeadline(time.Now().Add(5*time.Second)))
		wg.Add(2)
		go func() {
			defer wg.Done()
			write(conn1)
		}()
		read(conn1)
		read(conn1)
		read(conn1)
		go func() {
			defer wg.Done()
			write(conn1)
		}()
	}()
	wg.Wait()

	// about Deadline()
	require.NoError(t, conn1.SetDeadline(time.Now().Add(10*time.Millisecond)))
	require.NoError(t, conn2.SetDeadline(time.Now().Add(10*time.Millisecond)))
	time.Sleep(30 * time.Millisecond)
	buf := Bytes()
	_, err := conn1.Write(buf)
	require.Error(t, err)
	_, err = conn2.Read(buf)
	require.Error(t, err)

	require.NoError(t, conn1.SetDeadline(time.Now().Add(10*time.Millisecond)))
	require.NoError(t, conn2.SetDeadline(time.Now().Add(10*time.Millisecond)))
	time.Sleep(30 * time.Millisecond)
	buf = Bytes()
	_, err = conn1.Write(buf)
	require.Error(t, err)
	_, err = conn2.Read(buf)
	require.Error(t, err)

	// recover about net.Pipe()
	require.NoError(t, conn1.SetDeadline(time.Time{}))
	require.NoError(t, conn2.SetDeadline(time.Time{}))

	// Close()
	if close {
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

		// tls.Conn.Close still send data
		// so conn2 Close first
		require.NoError(t, conn2.Close())
		require.NoError(t, conn1.Close())
		wg.Wait()

		IsDestroyed(t, conn1)
		IsDestroyed(t, conn2)
	}
}
