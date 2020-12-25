package ftp

import (
	"bufio"
	"net"
	"strings"
	"time"

	"github.com/pkg/errors"

	"project/internal/nettool"

	"project/module/auxiliary/brute"
)

// reference:
// https://www.ietf.org/rfc/rfc0959.txt

const (
	codeServiceReady = "220"
	codeNeedPassword = "331"
	codeUserLoggedIn = "230"
	codeNotLoggedIn  = "530"

	cmdUser = "USER"
	cmdPass = "PASS"

	anonymousUser = "anonymous"
	anonymousPass = "IE@User"
)

type ftpConn struct {
	reader *bufio.Reader
	conn   net.Conn
}

func (fc *ftpConn) ReadResponse() (string, string, error) {
	line, _, err := fc.reader.ReadLine()
	if err != nil {
		return "", "", errors.Wrap(err, "failed to read response")
	}
	resp := strings.SplitN(string(line), " ", 2)
	if len(resp) != 2 {
		return "", "", errors.Errorf("read invalid response: %s", line)
	}
	return resp[0], resp[1], nil
}

func (fc *ftpConn) WriteRequest(cmd, arg string) error {
	// cmd + " " + arg + "\r\n"
	req := make([]byte, 0, len(cmd)+1+len(arg)+2)
	req = append(req, cmd...)
	req = append(req, ' ')
	req = append(req, arg...)
	req = append(req, "\r\n"...)
	_, err := fc.conn.Write(req)
	if err != nil {
		return errors.Wrap(err, "failed to write request")
	}
	return nil
}

func (fc *ftpConn) Close() error {
	return fc.conn.Close()
}

func connect(address string) (*ftpConn, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to connect target")
	}
	var ok bool
	defer func() {
		if !ok {
			_ = conn.Close()
		}
	}()
	conn = nettool.DeadlineConn(conn, time.Minute)
	reader := bufio.NewReader(conn)
	fc := ftpConn{reader: reader, conn: conn}
	code, arg, err := fc.ReadResponse()
	if err != nil {
		return nil, err
	}
	if code != codeServiceReady {
		const format = "connect with unexpected response code: %s, %s"
		return nil, errors.Errorf(format, code, arg)
	}
	ok = true
	return &fc, nil
}

func anonymousLogin(address string) (bool, error) {
	return login(address, anonymousUser, anonymousPass)
}

func login(address, username, password string) (bool, error) {
	conn, err := connect(address)
	if err != nil {
		return false, err
	}
	defer func() { _ = conn.Close() }()
	// write username
	err = conn.WriteRequest(cmdUser, username)
	if err != nil {
		return false, err
	}
	code, arg, err := conn.ReadResponse()
	if err != nil {
		return false, err
	}
	if code != codeNeedPassword {
		const format = "write username with unexpected response code: %s, %s"
		return false, errors.Errorf(format, code, arg)
	}
	// write password
	err = conn.WriteRequest(cmdPass, password)
	if err != nil {
		return false, err
	}
	code, arg, err = conn.ReadResponse()
	if err != nil {
		return false, err
	}
	// check response code
	switch code {
	case codeUserLoggedIn:
		return true, nil
	case codeNotLoggedIn:
		return false, brute.ErrInvalidCred
	default:
		const format = "write password with unexpected response code: %s, %s"
		return false, errors.Errorf(format, code, arg)
	}
}
