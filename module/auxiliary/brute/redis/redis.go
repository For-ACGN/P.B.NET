package redis

import (
	"fmt"
	"net"
	"time"
)

type redisConn struct {
	timeout time.Duration

	conn net.Conn
}

func (rc *redisConn) readSimpleStringsOrError() (string, error) {

	return "", nil

}

func connect(address, username, password string) (bool, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return false, err
	}
	defer func() { _ = conn.Close() }()

	// first try to send PING command, maybe redis server not require password
	ping := fmt.Sprintf("*1\r\n$4\r\nping\r\n")

	conn.Write([]byte(ping))

	conn.Read(nil)

	return false, nil
}
