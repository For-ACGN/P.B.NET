package redis

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"

	"github.com/pkg/errors"

	"project/internal/nettool"
	"project/internal/random"

	"project/module/auxiliary/brute"
)

func anonymousLogin(address string) (bool, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return false, errors.WithMessage(err, "failed to connect target")
	}
	defer func() { _ = conn.Close() }()
	conn = nettool.DeadlineConn(conn, time.Minute)
	// send a ping with random message
	msg := random.String(8 + random.Int(8))
	ping := []byte(fmt.Sprintf("*2\r\n$4\r\nping\r\n$%d\r\n%s\r\n", len(msg), msg))
	_, err = conn.Write(ping)
	if err != nil {
		return false, errors.Wrap(err, "failed to send ping message")
	}
	// read pong message
	reader := bufio.NewReader(conn)
	b, err := reader.ReadByte()
	if err != nil {
		return false, errors.Wrap(err, "failed to read pong")
	}
	// ok:   $8\r\n ran dom\r\n
	// fail: -NOAUTH Authentication require\r\n
	switch b {
	case '$':
		// read pong message
		line, _, err := reader.ReadLine()
		if err != nil {
			return false, errors.Wrap(err, "failed to read pong message length")
		}
		length, err := strconv.Atoi(string(line))
		if err != nil {
			return false, errors.Wrap(err, "invalid pong message length")
		}
		pong := make([]byte, length)
		_, err = io.ReadFull(reader, pong)
		if err != nil {
			return false, errors.Wrap(err, "failed to read pong message")
		}
		// compare with the generated random data
		if !bytes.Equal(pong, []byte(msg)) {
			return false, errors.Wrap(brute.ErrHoneypotDetected, "read invalid random data")
		}
		// read remaining text
		_, err = reader.ReadSlice('\n')
		return true, err
	case '-': // read remaining text
		_, err := reader.ReadSlice('\n')
		return false, err
	default:
		return false, errors.Wrapf(brute.ErrHoneypotDetected, "invalid pong: 0x%X", b)
	}
}

func login(address, username, password string) (bool, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return false, err
	}
	defer func() { _ = conn.Close() }()
	conn = nettool.DeadlineConn(conn, time.Minute)
	// send authenticate data
	var auth []byte
	if username == "" {
		const format = "*2\r\n$4\r\nauth\r\n$%d\r\n%s\r\n"
		auth = []byte(fmt.Sprintf(format, len(password), password))
	} else { // Redis 6.0 added ACL
		const format = "*3\r\n$4\r\nauth\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n"
		auth = []byte(fmt.Sprintf(format, len(username), username, len(password), password))
	}
	_, err = conn.Write(auth)
	if err != nil {
		return false, errors.Wrap(err, "failed to send ping message")
	}
	// read authenticate response
	reader := bufio.NewReader(conn)
	b, err := reader.ReadByte()
	if err != nil {
		return false, errors.Wrap(err, "failed to read authenticate response")
	}
	// ok:   +OK\r\n
	// fail: -ERR invalid password\r\n
	switch b {
	case '+':
		line, _, err := reader.ReadLine()
		if err != nil {
			return false, errors.Wrap(err, "failed to read authenticate response text")
		}
		if string(line) == "OK" {
			return true, nil
		}
		return false, errors.Wrap(brute.ErrHoneypotDetected, "invalid authenticate ok")
	case '-': // read remaining text
		_, err := reader.ReadSlice('\n')
		if err != nil {
			return false, errors.Wrap(err, "failed to read authenticate failed text")
		}
		return false, brute.ErrInvalidCred
	default:
		return false, errors.Wrapf(brute.ErrHoneypotDetected, "invalid authenticate response: 0x%X", b)
	}
}
