package nettool

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

// DialContext is a type alias about DialContext function.
type DialContext = func(ctx context.Context, network, address string) (net.Conn, error)

// ErrEmptyPort is an error of CheckPortString.
var ErrEmptyPort = errors.New("empty port")

// CheckPort is used to check port range.
func CheckPort(port int) error {
	if port < 0 || port > 65535 {
		return fmt.Errorf("invalid port: %d", port)
	}
	return nil
}

// CheckPortString is used to check port range, port is a string.
func CheckPortString(port string) error {
	if port == "" {
		return ErrEmptyPort
	}
	p, err := strconv.Atoi(port)
	if err != nil {
		return err
	}
	return CheckPort(p)
}

// JoinHostPort is used to join host and port to address.
func JoinHostPort(host string, port uint16) string {
	return net.JoinHostPort(host, strconv.Itoa(int(port)))
}

// SplitHostPort is used to split address to host and port.
func SplitHostPort(address string) (string, uint16, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return "", 0, err
	}
	portNum, err := strconv.Atoi(port)
	if err != nil {
		return "", 0, err
	}
	err = CheckPort(portNum)
	if err != nil {
		return "", 0, err
	}
	return host, uint16(portNum), nil
}

// IPToHost is used to convert IP address to URL.Host.
// net/http.Client need it(maybe it is a bug to handle IPv6 address when through proxy).
func IPToHost(address string) string {
	if !strings.Contains(address, ":") { // IPv4
		return address
	}
	return "[" + address + "]"
}

// IsTCPNetwork is used to check network is TCP.
func IsTCPNetwork(network string) error {
	switch network {
	case "tcp", "tcp4", "tcp6":
		return nil
	default:
		return fmt.Errorf("invalid tcp network: %s", network)
	}
}

// IsUDPNetwork is used to check network is UDP.
func IsUDPNetwork(network string) error {
	switch network {
	case "udp", "udp4", "udp6":
		return nil
	default:
		return fmt.Errorf("invalid udp network: %s", network)
	}
}

// IsNetClosingError is used to check this error is GOROOT/src/internal/poll.ErrNetClosing.
func IsNetClosingError(err error) bool {
	if err == nil {
		return false
	}
	const errStr = "use of closed network connection"
	return strings.Contains(err.Error(), errStr)
}

// EncodeExternalAddress is used to encode connection external address.
// If address is IP+Port, parse IP and return byte slice, ot return []byte(addr).
func EncodeExternalAddress(address string) []byte {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		// for special remote address
		return []byte(address)
	}
	ip := net.ParseIP(host)
	if ip != nil {
		return ip
	}
	// for special remote address
	return []byte(host)
}

// DecodeExternalAddress is used to decode connection external address.
// If address is a IP, return it, or return string(address).
func DecodeExternalAddress(address []byte) string {
	ip := net.IP(address).String()
	if strings.Contains(ip, ".") || strings.Contains(ip, ":") {
		return ip
	}
	return string(address)
}

// IPEnabled is used to get system IP enabled status.
func IPEnabled() (ipv4Enabled, ipv6Enabled bool) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return false, false
	}
	for _, iface := range interfaces {
		if iface.Flags != net.FlagUp|net.FlagBroadcast|net.FlagMulticast {
			continue
		}
		addresses, _ := iface.Addrs()
		for _, address := range addresses {
			ipAddr := strings.Split(address.String(), "/")[0]
			ip := net.ParseIP(ipAddr)
			ip4 := ip.To4()
			if ip4 != nil {
				if ip4.IsGlobalUnicast() {
					ipv4Enabled = true
				}
			} else {
				if ip.To16().IsGlobalUnicast() {
					ipv6Enabled = true
				}
			}
			if ipv4Enabled && ipv6Enabled {
				break
			}
		}
	}
	return
}

type deadlineConn struct {
	net.Conn
	deadline time.Duration
}

func (d *deadlineConn) Read(p []byte) (n int, err error) {
	_ = d.Conn.SetReadDeadline(time.Now().Add(d.deadline))
	return d.Conn.Read(p)
}

func (d *deadlineConn) Write(p []byte) (n int, err error) {
	_ = d.Conn.SetWriteDeadline(time.Now().Add(d.deadline))
	return d.Conn.Write(p)
}

// DeadlineConn is used to return a net.Conn that SetReadDeadline()
// and SetWriteDeadline() before each Read() and Write().
func DeadlineConn(conn net.Conn, deadline time.Duration) net.Conn {
	dc := deadlineConn{
		Conn:     conn,
		deadline: deadline,
	}
	if dc.deadline < 1 {
		dc.deadline = time.Minute
	}
	return &dc
}

// FprintConn is used to print information about net.Conn to a io.Writer.
//
// local:  tcp 127.0.0.1:1234
// remote: tcp 127.0.0.1:1235
func FprintConn(w io.Writer, c net.Conn) {
	_, _ = fmt.Fprintf(w, "local:  %s %s\nremote: %s %s",
		c.LocalAddr().Network(), c.LocalAddr(),
		c.RemoteAddr().Network(), c.RemoteAddr(),
	)
}

// PrintConn is used to print information about net.Conn to a *bytes.Buffer.
func PrintConn(conn net.Conn) *bytes.Buffer {
	buf := bytes.NewBuffer(make([]byte, 0, 64))
	FprintConn(buf, conn)
	return buf
}

// Server implemented method Addresses() []net.Addr.
type Server interface {
	Addresses() []net.Addr
}

// WaitServerServe is used to wait server serve.
func WaitServerServe(ctx context.Context, errCh <-chan error, srv Server, n int) ([]net.Addr, error) {
	if n < 1 {
		panic("n < 1")
	}
	timer := time.NewTicker(25 * time.Millisecond)
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			addrs := srv.Addresses()
			if len(addrs) >= n {
				return addrs, nil
			}
		case err := <-errCh:
			if err != nil {
				return nil, err
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}
