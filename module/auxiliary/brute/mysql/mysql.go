package mysql

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"time"

	"github.com/pkg/errors"

	"project/module/auxiliary/brute"
)

// Brute is the brute module.
type Brute struct {
	*brute.Brute
}

// Name is used to get brute module name.
func (b *Brute) Name() string {
	return "MySQL Brute"
}

// Description is used to get brute module description.
func (b *Brute) Description() string {
	return "MySQL Brute"
}

// Config contains configuration about mysql brute module.
type Config struct {
}

func connect(address, username, password string) (bool, error) {
	conn, err := net.Dial("tcp", address) // TODO set proxy
	if err != nil {
		return false, err
	}
	defer func() { _ = conn.Close() }()
	mc := mysqlConn{
		username: username,
		password: password,
		timeout:  time.Minute, // TODO set able
		buf:      bytes.NewBuffer(make([]byte, 0, 128)),
		conn:     conn,
	}
	authData, plugin, err := mc.readGreetingPacket()
	if err != nil {
		return false, err
	}
	err = mc.writeLoginRequest(authData, plugin)
	if err != nil {
		return false, err
	}
	err = mc.handleAuthResult(authData, plugin)
	if err != nil {
		return false, err
	}

	return true, nil
}

// mError is an error type which represents a single MySQL error.
type mError struct {
	Number  uint16
	Message string
}

func (e *mError) Error() string {
	return fmt.Sprintf("Error %d: %s", e.Number, e.Message)
}

func handleErrorPacket(data []byte) error {
	if data[0] != iERR {
		return errors.New("malformed packet")
	}
	// 0xff [1 byte], Error Number [16 bit uint]
	errno := binary.LittleEndian.Uint16(data[1:3])
	// 1792: ER_CANT_EXECUTE_IN_READ_ONLY_TRANSACTION
	// 1290: ER_OPTION_PREVENTS_STATEMENT (returned by Aurora during fail over)
	if errno == 1792 || errno == 1290 {
		// Oops; we are connected to a read-only connection, and won't be able
		// to issue any write statements. Since RejectReadOnly is configured,
		// we throw away this connection hoping this one would have write
		// permission. This is specifically for a possible race condition
		// during fail over (e.g. on AWS Aurora). See README.md for more.
		//
		// We explicitly close the connection before returning
		// driver.ErrBadConn to ensure that `database/sql` purges this
		// connection and initiates a new one for next statement next time.
		return errors.New("bad connection")
	}
	pos := 3
	// SQL State [optional: # + 5bytes string]
	if data[3] == 0x23 {
		// sqlstate := string(data[4 : 4+5])
		pos = 9
	}
	// Error Message [string]
	return &mError{Number: errno, Message: string(data[pos:])}
}
