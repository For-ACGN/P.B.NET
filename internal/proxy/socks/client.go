package socks

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/pkg/errors"

	"project/internal/nettool"
	"project/internal/security"
	"project/internal/xpanic"
)

// Client implemented internal/proxy.client.
type Client struct {
	network    string
	address    string
	socks4     bool
	disableExt bool   // socks4 can't resolve domain name
	protocol   string // "socks5", "socks4a", "socks4"

	// options
	username *security.String
	password *security.String
	userID   *security.String
	timeout  time.Duration
}

// NewSocks5Client is used to create a socks5 client.
func NewSocks5Client(network, address string, opts *Options) (*Client, error) {
	return newClient(network, address, opts, false, false)
}

// NewSocks4aClient is used to create a socks4a client.
func NewSocks4aClient(network, address string, opts *Options) (*Client, error) {
	return newClient(network, address, opts, true, false)
}

// NewSocks4Client is used to create a socks4 client.
func NewSocks4Client(network, address string, opts *Options) (*Client, error) {
	return newClient(network, address, opts, true, true)
}

func newClient(network, address string, opts *Options, socks4, disableExt bool) (*Client, error) {
	err := CheckNetworkAndAddress(network, address)
	if err != nil {
		return nil, err
	}
	if opts == nil {
		opts = new(Options)
	}
	client := Client{
		network:    network,
		address:    address,
		socks4:     socks4,
		disableExt: disableExt,
		timeout:    opts.Timeout,
	}
	// select protocol
	switch {
	case !socks4:
		client.protocol = "socks5"
	case socks4 && disableExt:
		client.protocol = "socks4"
	case socks4 && !disableExt:
		client.protocol = "socks4a"
	}
	// set options
	if opts.Username != "" || opts.Password != "" {
		client.username = security.NewString(opts.Username)
		client.password = security.NewString(opts.Password)
	}
	if opts.UserID != "" {
		client.userID = security.NewString(opts.UserID)
	}
	if client.timeout < 1 {
		client.timeout = defaultDialTimeout
	}
	return &client, nil
}

// Dial is used to connect to address through proxy.
func (c *Client) Dial(network, address string) (net.Conn, error) {
	err := CheckNetworkAndAddress(network, address)
	if err != nil {
		const format = "dial: %s client %s connect %s with error: %s"
		return nil, errors.Errorf(format, c.protocol, c.address, address, err)
	}
	conn, err := (&net.Dialer{Timeout: c.timeout}).Dial(c.network, c.address)
	if err != nil {
		const format = "dial: failed to connect %s server %s"
		return nil, errors.Wrapf(err, format, c.protocol, c.address)
	}
	_, err = c.Connect(context.Background(), conn, network, address)
	if err != nil {
		_ = conn.Close()
		const format = "dial: %s client %s failed to connect %s"
		return nil, errors.WithMessagef(err, format, c.protocol, c.address, address)
	}
	_ = conn.SetDeadline(time.Time{})
	return conn, nil
}

// DialContext is used to connect to address through proxy with context.
func (c *Client) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	err := CheckNetworkAndAddress(network, address)
	if err != nil {
		const format = "dial context: %s client %s connect %s with error: %s"
		return nil, errors.Errorf(format, c.protocol, c.address, address, err)
	}
	conn, err := (&net.Dialer{Timeout: c.timeout}).DialContext(ctx, c.network, c.address)
	if err != nil {
		const format = "dial context: failed to connect %s server %s"
		return nil, errors.Wrapf(err, format, c.protocol, c.address)
	}
	_, err = c.Connect(ctx, conn, network, address)
	if err != nil {
		_ = conn.Close()
		const format = "dial context: %s client %s failed to connect %s"
		return nil, errors.WithMessagef(err, format, c.protocol, c.address, address)
	}
	_ = conn.SetDeadline(time.Time{})
	return conn, nil
}

// DialTimeout is used to connect to address through proxy with timeout.
func (c *Client) DialTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
	err := CheckNetworkAndAddress(network, address)
	if err != nil {
		const format = "dial timeout: %s client %s connect %s with error: %s"
		return nil, errors.Errorf(format, c.protocol, c.address, address, err)
	}
	if timeout < 1 {
		timeout = defaultDialTimeout
	}
	conn, err := (&net.Dialer{Timeout: timeout}).Dial(c.network, c.address)
	if err != nil {
		const format = "dial timeout: failed to connect %s server %s"
		return nil, errors.Wrapf(err, format, c.protocol, c.address)
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	_, err = c.Connect(ctx, conn, network, address)
	if err != nil {
		_ = conn.Close()
		const format = "dial timeout: %s client %s failed to connect %s"
		return nil, errors.WithMessagef(err, format, c.protocol, c.address, address)
	}
	_ = conn.SetDeadline(time.Time{})
	return conn, nil
}

// Connect is used to connect to address through proxy with context.
func (c *Client) Connect(ctx context.Context, conn net.Conn, network, address string) (net.Conn, error) {
	err := CheckNetworkAndAddress(network, address)
	if err != nil {
		return nil, err
	}
	host, port, err := nettool.SplitHostPort(address)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	_ = conn.SetDeadline(time.Now().Add(c.timeout))
	// interrupt
	var errCh chan error
	if ctx.Done() != nil {
		errCh = make(chan error, 2)
	}
	if errCh == nil {
		if c.socks4 {
			err = c.connectSocks4(conn, host, port)
		} else {
			err = c.connectSocks5(conn, host, port)
		}
	} else {
		go func() {
			defer close(errCh)
			defer func() {
				if r := recover(); r != nil {
					buf := xpanic.Log(r, "Client.Connect")
					errCh <- fmt.Errorf(buf.String())
				}
			}()
			if c.socks4 {
				errCh <- c.connectSocks4(conn, host, port)
			} else {
				errCh <- c.connectSocks5(conn, host, port)
			}
		}()
		select {
		case err = <-errCh:
			if err != nil {
				// if the error was due to the context
				// closing, prefer the context's error, rather
				// than some random network teardown error.
				if e := ctx.Err(); e != nil {
					err = e
				}
			}
		case <-ctx.Done():
			err = ctx.Err()
		}
	}
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	_ = conn.SetDeadline(time.Time{})
	return conn, nil
}

// HTTP is used to set *http.Transport about proxy.
func (c *Client) HTTP(t *http.Transport) {
	t.DialContext = c.DialContext
}

// Timeout is used to get the socks client timeout.
func (c *Client) Timeout() time.Duration {
	return c.timeout
}

// Server is used to get the socks server address.
func (c *Client) Server() (string, string) {
	return c.network, c.address
}

// Info is used to get the socks client information.
//
// socks5, server: tcp 127.0.0.1:1080
// socks5, server: tcp 127.0.0.1:1080, auth: admin:123456
// socks4a, server: tcp 127.0.0.1:1080, user id: test
func (c *Client) Info() string {
	buf := new(bytes.Buffer)
	_, _ = fmt.Fprintf(buf, "%s, server: %s %s", c.protocol, c.network, c.address)
	if c.socks4 {
		if c.userID != nil {
			userID := c.userID.Get()
			defer c.userID.Put(userID)
			_, _ = fmt.Fprintf(buf, ", user id: %s", userID)
		}
	} else {
		if c.username != nil {
			username := c.username.Get()
			defer c.username.Put(username)
			password := c.password.Get()
			defer c.password.Put(password)
			_, _ = fmt.Fprintf(buf, ", auth: %s:%s", username, password)
		}
	}
	return buf.String()
}
