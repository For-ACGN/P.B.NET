package mysql

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"

	"project/internal/convert"
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
		DBName:                  "mysql",
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
	timeFormat         = "2006-01-02 15:04:05.999999"
)

// MySQL constants documentation:
// http://dev.mysql.com/doc/internals/en/client-server-protocol.html

const (
	iOK           byte = 0x00
	iAuthMoreData byte = 0x01
	iLocalInFile  byte = 0xfb
	iEOF          byte = 0xfe
	iERR          byte = 0xff
)

const (
	cachingSha2PasswordRequestPublicKey          = 2
	cachingSha2PasswordFastAuthSuccess           = 3
	cachingSha2PasswordPerformFullAuthentication = 4
)

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

// +---------------+---------------+
// | Packet Length | Packet Number |
// +---------------+---------------+
// |  3 bytes(LE)  |    1 byte     |
// +---------------+---------------+

type mysqlConn struct {
	buf     *bytes.Buffer
	conn    net.Conn
	timeout time.Duration
}

func (mc *mysqlConn) ReadPacket() ([]byte, uint8, error) {
	header := make([]byte, 4)
	_ = mc.conn.SetDeadline(time.Now().Add(mc.timeout))
	_, err := io.ReadFull(mc.conn, header)
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to read packet header")
	}
	packetLen := int64(uint32(header[0]) | uint32(header[1])<<8 | uint32(header[2])<<16)
	packetNum := header[3]
	// read packet data
	_, err = io.CopyN(mc.buf, mc.conn, packetLen)
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to read packet data")
	}
	return mc.buf.Bytes(), packetNum, err
}

func (mc *mysqlConn) WritePacket(data []byte, num uint8) error {
	return nil
}

// read will read data until reach 0x00.
func readUntilNull(reader io.Reader) ([]byte, error) {
	data := make([]byte, 0, 8)
	buf := make([]byte, 1)
	for {
		n, err := reader.Read(buf)
		if err != nil {
			return nil, err
		}
		data = append(data, buf[:n]...)
		if n != 0 && buf[0] == 0x00 {
			return data[:len(data)-1], nil
		}
	}
}

// Greeting is the mysql server greeting, Cap is capabilities.
type Greeting struct {
	Protocol      uint8
	Version       string
	ThreadID      uint32
	SaltFirst     []byte
	ServerCap     uint16
	ServerLang    uint8
	ServerStatus  uint16
	ExtServerCap  uint16
	AuthPluginLen uint8
	unused        [10]byte
	AuthPlugin    string
}

func parseServerGreeting(packet []byte) (*Greeting, error) {
	if len(packet) < 1 {
		return nil, errors.New("malformed packet")
	}
	if packet[0] == iERR {
		return nil, handleErrorPacket(packet)
	}
	// protocol
	protocol := packet[0]
	if protocol < minProtocolVersion {
		const format = "unsupported protocol version %d. Version %d or higher is required"
		return nil, fmt.Errorf(format, protocol, minProtocolVersion)
	}
	reader := bytes.NewReader(packet[1:])
	// version
	version, err := readUntilNull(reader)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read version")
	}
	threadID := make([]byte, 0, 4)
	_, err = io.ReadFull(reader, threadID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read thread id")
	}
	saltFirst, err := readUntilNull(reader)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read salt about first party")
	}
	serverCapabilities := make([]byte, 2)
	_, err = io.ReadFull(reader, serverCapabilities)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read server capabilities")
	}

	greeting := Greeting{
		Protocol:  protocol,
		Version:   string(version),
		ThreadID:  convert.LEBytesToUint32(threadID),
		SaltFirst: saltFirst,
		ServerCap: convert.LEBytesToUint16(serverCapabilities),
	}
	return &greeting, nil
}

func connect(address string, username, password string) (bool, error) {
	// TODO set proxy
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return false, err
	}
	defer func() { _ = conn.Close() }()

	mc := mysqlConn{
		buf:     bytes.NewBuffer(make([]byte, 0, 128)),
		conn:    conn,
		timeout: time.Minute, // TODO set able
	}
	packet, num, err := mc.ReadPacket()
	if err != nil {
		return false, errors.WithMessage(err, "failed to read server greeting")
	}

	greeting, err := parseServerGreeting(packet)
	if err != nil {
		return false, err
	}

	fmt.Println(greeting.AuthPlugin, num, err)
	return true, nil
}
