package dns

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"

	"project/internal/convert"
	"project/internal/xnet/xnetutil"
)

const (
	defaultTimeout     = 10 * time.Second // udp is 5 second
	defaultMaxBodySize = 65535            // about DOH
	headerSize         = 2                // tcp && tls need it
)

// ErrNoConnection is an error of the dial
var ErrNoConnection = fmt.Errorf("no connection")

func resolve(ctx context.Context, address, domain string, opts *Options) ([]string, error) {
	message := packMessage(types[opts.Type], domain)
	var err error
	switch opts.Method {
	case MethodUDP:
		message, err = dialUDP(ctx, address, message, opts)
	case MethodTCP:
		message, err = dialTCP(ctx, address, message, opts)
	case MethodDoT:
		message, err = dialDoT(ctx, address, message, opts)
	case MethodDoH:
		message, err = dialDoH(ctx, address, message, opts)
	}
	if err != nil {
		return nil, err
	}
	return unpackMessage(message)
}

// if question > 512 use tcp tls doh
func dialUDP(ctx context.Context, address string, message []byte, opts *Options) ([]byte, error) {
	network := opts.Network
	switch network {
	case "":
		network = "udp"
	case "udp", "udp4", "udp6":
	default:
		return nil, errors.WithStack(net.UnknownNetworkError(network))
	}
	// set timeout
	timeout := opts.Timeout
	if timeout < 1 {
		timeout = 5 * time.Second
	}
	// dial
	for i := 0; i < 3; i++ {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		conn, err := opts.dialContext(ctx, network, address)
		if err != nil {
			cancel()
			return nil, errors.WithStack(err) // not continue
		}
		dConn := xnetutil.DeadlineConn(conn, timeout)
		_, _ = dConn.Write(message)
		buffer := make([]byte, 512)
		n, err := dConn.Read(buffer)
		if err == nil {
			_ = dConn.Close()
			cancel()
			return buffer[:n], nil
		}
		_ = dConn.Close()
		cancel()
	}
	return nil, errors.WithStack(ErrNoConnection)
}

func sendMessage(conn net.Conn, message []byte, timeout time.Duration) ([]byte, error) {
	dConn := xnetutil.DeadlineConn(conn, timeout)
	defer func() { _ = dConn.Close() }()
	// add size header
	header := new(bytes.Buffer)
	header.Write(convert.Uint16ToBytes(uint16(len(message))))
	header.Write(message)
	_, err := dConn.Write(header.Bytes())
	if err != nil {
		return nil, errors.WithStack(err)
	}
	// read message size
	length := make([]byte, headerSize)
	_, err = io.ReadFull(dConn, length)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	resp := make([]byte, int(convert.BytesToUint16(length)))
	_, err = io.ReadFull(dConn, resp)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return resp, nil
}

func dialTCP(ctx context.Context, address string, message []byte, opts *Options) ([]byte, error) {
	network := opts.Network
	switch network {
	case "":
		network = "tcp" // default
	case "tcp", "tcp4", "tcp6":
	default:
		return nil, errors.WithStack(net.UnknownNetworkError(network))
	}
	// set timeout
	timeout := opts.Timeout
	if timeout < 1 {
		timeout = defaultTimeout
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	// dial
	conn, err := opts.dialContext(ctx, network, address)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return sendMessage(conn, message, timeout)
}

func dialDoT(ctx context.Context, config string, message []byte, opts *Options) ([]byte, error) {
	network := opts.Network
	switch network {
	case "": // default
		network = "tcp"
	case "tcp", "tcp4", "tcp6":
	default:
		return nil, errors.WithStack(net.UnknownNetworkError(network))
	}
	// load configs
	configs := strings.Split(config, "|")
	host, port, err := net.SplitHostPort(configs[0])
	if err != nil {
		return nil, errors.WithStack(err)
	}
	// set TLS Config
	tlsConfig, err := opts.TLSConfig.Apply()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if tlsConfig.ServerName == "" {
		tlsConfig.ServerName = host
	}
	// set timeout
	timeout := opts.Timeout
	if timeout < 1 {
		timeout = 2 * defaultTimeout
	}
	var conn *tls.Conn
	switch len(configs) {
	case 1: // ip mode
		// 8.8.8.8:853
		// [2606:4700:4700::1001]:853
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		rawConn, err := opts.dialContext(ctx, network, config)
		if err != nil {
			return nil, err
		}
		conn = tls.Client(rawConn, tlsConfig)
	case 2: // domain mode
		// dns.google:853|8.8.8.8,8.8.4.4
		// cloudflare-dns.com:853|2606:4700:4700::1001,2606:4700:4700::1111
		ips := strings.Split(strings.TrimSpace(configs[1]), ",")
		var rawConn net.Conn
		for i := 0; i < len(ips); i++ {
			ctx, cancel := context.WithTimeout(ctx, timeout)
			rawConn, err = opts.dialContext(ctx, network, net.JoinHostPort(ips[i], port))
			if err == nil {
				conn = tls.Client(rawConn, tlsConfig)
				cancel()
				break
			}
			cancel()
		}
		if err != nil {
			return nil, errors.Wrap(ErrNoConnection, err.Error())
		}
	default:
		return nil, errors.Errorf("invalid config: %s", config)
	}
	return sendMessage(conn, message, timeout)
}

// support RFC 8484
func dialDoH(ctx context.Context, server string, question []byte, opts *Options) ([]byte, error) {
	str := base64.RawURLEncoding.EncodeToString(question)
	url := fmt.Sprintf("%s?ct=application/dns-message&dns=%s", server, str)
	var (
		req *http.Request
		err error
	)
	if len(url) < 2048 { // GET
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	} else { // POST
		body := bytes.NewReader(question)
		req, err = http.NewRequestWithContext(ctx, http.MethodPost, server, body)
	}
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// set header
	req.Header = opts.Header.Clone()
	if req.Header == nil {
		req.Header = make(http.Header)
	}
	if req.Method == http.MethodPost {
		req.Header.Set("Content-Type", "application/dns-message")
	}
	req.Header.Set("Accept", "application/dns-message")

	// http client
	client := http.Client{
		Transport: opts.transport,
		Timeout:   opts.Timeout,
	}
	defer client.CloseIdleConnections()
	if client.Timeout < 1 {
		client.Timeout = 2 * defaultTimeout
	}
	maxBodySize := opts.MaxBodySize
	if maxBodySize < 1 {
		maxBodySize = defaultMaxBodySize
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer func() { _ = resp.Body.Close() }()
	return ioutil.ReadAll(io.LimitReader(resp.Body, maxBodySize))
}
