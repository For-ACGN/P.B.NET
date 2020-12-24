package ftp

import (
	"bufio"
	"net"
	"strings"
	"time"

	"github.com/pkg/errors"

	"project/internal/nettool"
)

func connect(address string) (*bufio.Reader, net.Conn, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "failed to connect target")
	}
	conn = nettool.DeadlineConn(conn, time.Minute)
	reader := bufio.NewReader(conn)
	line, _, err := reader.ReadLine()
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to read service ready")
	}
	resp := strings.SplitN(string(line), " ", 2)
	if len(resp) != 2 {
		return nil, nil, errors.Errorf("read invalid response: %s", line)
	}
	if resp[0] != "220" {
		return nil, nil, errors.Errorf("unexcepted response code: %s", resp[0])
	}
	return reader, conn, nil
}

func anonymousLogin(address string) (bool, error) {
	reader, conn, err := connect(address)
	if err != nil {
		return false, err
	}
	defer func() { _ = conn.Close() }()

	reader.ReadLine()
	return true, nil
}

func login(address, username, password string) (bool, error) {
	reader, conn, err := connect(address)
	if err != nil {
		return false, err
	}
	defer func() { _ = conn.Close() }()

	reader.ReadLine()
	return true, nil
}
