package socks

import (
	"bytes"
	"context"
	"crypto/subtle"
	"io"
	"net"
	"strconv"

	"github.com/pkg/errors"

	"project/internal/convert"
	"project/internal/logger"
	"project/internal/nettool"
)

// http://ftp.icm.edu.pl/packages/socks/socks4/SOCKS4.protocol
// http://www.openssh.com/txt/socks4a.protocol

const (
	version4    = 0x04
	v4Succeeded = 0x5a
	v4Refused   = 0x5b
	v4Ident     = 0x5c
	v4InvalidID = 0x5d
)

var v4IPPadding = []byte{0x00, 0x00, 0x00, 0x01} // domain

type v4Reply uint8

func (r v4Reply) String() string {
	switch r {
	case v4Refused:
		return "request rejected or failed"
	case v4Ident:
		return "request rejected because SOCKS server cannot connect to ident on the client"
	case v4InvalidID:
		return "request rejected because the client program and ident report different user-ids"
	default:
		return "unknown reply: " + strconv.Itoa(int(r))
	}
}

func (c *Client) connectSocks4(conn net.Conn, host string, port uint16) error {
	var (
		hostData   []byte
		socks4aExt bool
	)
	ip := net.ParseIP(host)
	if ip != nil {
		ip4 := ip.To4()
		if ip4 != nil {
			hostData = ip4
		} else {
			return errors.New("socks4 or socks4a server don't support IPv6")
		}
	} else if c.disableExt {
		const format = "%s is a socks4 server and don't support hostname"
		return errors.Errorf(format, c.address)
	} else {
		l := len(host)
		if l > 255 {
			return errors.New("hostname too long")
		}
		hostData = make([]byte, l)
		copy(hostData, host)
		socks4aExt = true
	}

	// handshake
	buffer := bytes.Buffer{}
	buffer.WriteByte(version4)
	buffer.WriteByte(connect)
	buffer.Write(convert.BEUint16ToBytes(port))
	if socks4aExt { // socks4a ext
		buffer.Write(v4IPPadding) // padding IPv4
	} else {
		buffer.Write(hostData) // IPv4
	}
	// user id
	buffer.Write(c.userID)
	buffer.WriteByte(0x00) // NULL
	// write domain
	if socks4aExt {
		buffer.Write(hostData)
		buffer.WriteByte(0x00) // NULL
	}
	_, err := conn.Write(buffer.Bytes())
	if err != nil {
		return errors.Wrap(err, "failed to write socks4 request data")
	}

	// read response, version4, reply, port, ip
	reply := make([]byte, 1+1+2+net.IPv4len)
	_, err = io.ReadFull(conn, reply)
	if err != nil {
		return errors.Wrap(err, "failed to read socks4 reply")
	}
	if reply[0] != 0x00 { // must 0x00 not 0x04
		return errors.Errorf("invalid socks version %d", reply[0])
	}
	if reply[1] != v4Succeeded {
		return errors.New(v4Reply(reply[1]).String())
	}
	return nil
}

var (
	v4ReplySucceeded = []byte{0x00, v4Succeeded, 0, 0, 0, 0, 0, 0}
	v4ReplyRefused   = []byte{0x00, v4Refused, 0, 0, 0, 0, 0, 0}
)

func (c *conn) serveSocks4() {
	// 10 = version(1) + cmd(1) + port(2) + address(4) + 2xNULL(2) maybe
	// 16 = domain name
	buf := make([]byte, 10+16) // prepare
	_, err := io.ReadFull(c.local, buf[:8])
	if err != nil {
		c.log(logger.Error, "failed to read socks4 request:", err)
		return
	}
	// check version
	if buf[0] != version4 {
		c.log(logger.Error, "unexpected socks4 version")
		return
	}
	// command
	if buf[1] != connect {
		c.log(logger.Error, "unknown command:", buf[1])
		return
	}
	if !c.checkUserID() {
		return
	}
	// address
	port := convert.BEBytesToUint16(buf[2:4])
	var (
		domain bool
		ip     bool
		host   string
	)
	if c.server.disableExt {
		ip = true
	} else {
		// check is domain, 0.0.0.x is domain mode
		if bytes.Equal(buf[4:7], []byte{0x00, 0x00, 0x00}) && buf[7] != 0x00 {
			domain = true
		} else {
			ip = true
		}
	}
	if ip {
		host = net.IPv4(buf[4], buf[5], buf[6], buf[7]).String()
	}
	if domain { // read domain
		var domainName []byte
		for {
			_, err = c.local.Read(buf[:1])
			if err != nil {
				c.log(logger.Error, "failed to read domain name:", err)
				return
			}
			// find 0x00(end)
			if buf[0] == 0x00 {
				break
			}
			domainName = append(domainName, buf[0])
		}
		host = string(domainName)
	}
	address := nettool.JoinHostPort(host, port)
	// connect target
	c.log(logger.Info, "connect:", address)
	ctx, cancel := context.WithTimeout(c.server.ctx, c.server.timeout)
	defer cancel()
	remote, err := c.server.dialContext(ctx, "tcp", address)
	if err != nil {
		c.log(logger.Error, "failed to connect target:", err)
		_, _ = c.local.Write(v4ReplyRefused)
		return
	}
	// write reply
	_, err = c.local.Write(v4ReplySucceeded)
	if err != nil {
		c.log(logger.Error, "failed to write reply", err)
		_ = remote.Close()
		return
	}
	c.remote = remote
}

func (c *conn) checkUserID() bool {
	var (
		userID []byte
		err    error
	)
	buffer := make([]byte, 1)
	for {
		_, err = c.local.Read(buffer)
		if err != nil {
			c.log(logger.Error, "failed to read user id:", err)
			return false
		}
		// find 0x00(end)
		if buffer[0] == 0x00 {
			break
		}
		userID = append(userID, buffer[0])
	}
	// compare user id
	if subtle.ConstantTimeCompare(c.server.userID, userID) != 1 {
		c.logf(logger.Exploit, "invalid user id: %s", userID)
		return false
	}
	return true
}
