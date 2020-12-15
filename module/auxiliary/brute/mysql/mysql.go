package mysql

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
)

// Brute is
func Brute(address string, usernames, passwords []string) (string, string, bool) {
	for _, username := range usernames {
		for _, password := range passwords {
			if Login(address, username, password) {
				return username, password, true
			}
		}
	}
	return "", "", false
}

// Login is
func Login(address string, username, password string) bool {
	connector, err := mysql.NewConnector(&mysql.Config{
		User:                    username,
		Passwd:                  password,
		Addr:                    address,
		DBName:                  "",
		Collation:               "utf8mb4_general_ci",
		AllowCleartextPasswords: true,
		AllowNativePasswords:    true,
		AllowOldPasswords:       true,
	})
	if err != nil {
		// fmt.Println("1", err)
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	conn, err := connector.Connect(ctx)
	if err != nil {
		return false
	}
	defer func() { _ = conn.Close() }()
	return true
}

const (
	defaultAuthPlugin  = "mysql_native_password"
	minProtocolVersion = 10
	maxPacketSize      = 1<<24 - 1
)

// MySQL constants documentation:
// http://dev.mysql.com/doc/internals/en/client-server-protocol.html

const (
	iOK           byte = 0x00
	iAuthMoreData byte = 0x01
	iEOF          byte = 0xfe
	iERR          byte = 0xff
)

const (
	cachingSha2PasswordRequestPublicKey          = 2
	cachingSha2PasswordFastAuthSuccess           = 3
	cachingSha2PasswordPerformFullAuthentication = 4
)

func connect(address string, username, password string) (bool, error) {
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

// Error is an error type which represents a single MySQL error.
type Error struct {
	Number  uint16
	Message string
}

func (e *Error) Error() string {
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
	return &Error{
		Number:  errno,
		Message: string(data[pos:]),
	}
}
