package mysql

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"time"

	"github.com/pkg/errors"

	"project/internal/convert"
)

// +---------------+---------------+
// | Packet Length | Packet Number |
// +---------------+---------------+
// |  3 bytes(LE)  |    1 byte     |
// +---------------+---------------+

type mysqlConn struct {
	username string
	password string
	timeout  time.Duration

	buf    *bytes.Buffer
	conn   net.Conn
	pktNum uint8
}

func (mc *mysqlConn) readPacket() ([]byte, error) {
	header := make([]byte, 4)
	_ = mc.conn.SetReadDeadline(time.Now().Add(mc.timeout))
	_, err := io.ReadFull(mc.conn, header)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read packet header")
	}
	packetLen := int64(uint32(header[0]) | uint32(header[1])<<8 | uint32(header[2])<<16)
	// read packet data
	mc.buf.Reset()
	_, err = io.CopyN(mc.buf, mc.conn, packetLen)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read packet data")
	}
	mc.pktNum = header[3] // packet number
	return mc.buf.Bytes(), err
}

func (mc *mysqlConn) writePacket(packet []byte) error {
	header := convert.LEUint32ToBytes(uint32(len(packet)))
	header[3] = mc.pktNum + 1
	_ = mc.conn.SetWriteDeadline(time.Now().Add(mc.timeout))
	_, err := mc.conn.Write(append(header, packet...))
	if err != nil {
		return errors.Wrap(err, "failed to write packet")
	}
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

// Greeting is the MySQL server greeting, Cap is capabilities.
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
	SaltSecond    []byte
	AuthPlugin    string
}

func parseGreeting(packet []byte) (*Greeting, error) {
	if packet[0] == iERR {
		return nil, handleErrorPacket(packet)
	}
	// protocol
	protocol := packet[0]
	if protocol < minProtocolVersion {
		const format = "unsupported protocol version %d. version %d or higher is required"
		return nil, fmt.Errorf(format, protocol, minProtocolVersion)
	}
	reader := bytes.NewReader(packet[1:])
	// version
	version, err := readUntilNull(reader)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read version")
	}
	// thread id
	threadID := make([]byte, 4)
	_, err = io.ReadFull(reader, threadID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read thread id")
	}
	// salt first part
	saltFirst, err := readUntilNull(reader)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read salt about first part")
	}
	// server capabilities
	serverCapabilities := make([]byte, 2)
	_, err = io.ReadFull(reader, serverCapabilities)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read server capabilities")
	}
	// server language
	serverLanguage := make([]byte, 1)
	_, err = io.ReadFull(reader, serverLanguage)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read server language")
	}
	// server status
	serverStatus := make([]byte, 2)
	_, err = io.ReadFull(reader, serverStatus)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read server status")
	}
	// extended server capabilities
	extendedServerCapabilities := make([]byte, 2)
	_, err = io.ReadFull(reader, extendedServerCapabilities)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read extended server capabilities")
	}
	// authentication plugin length
	authPluginLength := make([]byte, 1)
	_, err = io.ReadFull(reader, authPluginLength)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read authentication plugin length")
	}
	// unused data
	_, err = io.CopyN(ioutil.Discard, reader, 10)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read unused data")
	}
	//  salt second part
	saltSecond, err := readUntilNull(reader)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read salt about second part")
	}
	// authentication plugin
	authPlugin, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read authentication plugin")
	}
	end := bytes.IndexByte(authPlugin, 0x00)
	if end != -1 {
		authPlugin = authPlugin[:end]
	}
	greeting := Greeting{
		Protocol:      protocol,
		Version:       string(version),
		ThreadID:      convert.LEBytesToUint32(threadID),
		SaltFirst:     saltFirst,
		ServerCap:     convert.LEBytesToUint16(serverCapabilities),
		ServerLang:    serverLanguage[0],
		ServerStatus:  convert.LEBytesToUint16(serverStatus),
		ExtServerCap:  convert.LEBytesToUint16(extendedServerCapabilities),
		AuthPluginLen: authPluginLength[0],
		SaltSecond:    saltSecond,
		AuthPlugin:    string(authPlugin),
	}
	return &greeting, nil
}

func (mc *mysqlConn) readGreetingPacket() ([]byte, string, error) {
	packet, err := mc.readPacket()
	if err != nil {
		return nil, "", errors.WithMessage(err, "failed to read server greeting")
	}
	if len(packet) < 1 {
		return nil, "", errors.New("malformed greeting packet")
	}
	greeting, err := parseGreeting(packet)
	if err != nil {
		return nil, "", err
	}
	authData := make([]byte, len(greeting.SaltFirst)+len(greeting.SaltSecond))
	copy(authData, greeting.SaltFirst)
	copy(authData[len(greeting.SaltFirst):], greeting.SaltSecond)
	plugin := greeting.AuthPlugin
	if plugin == "" {
		plugin = defaultAuthPlugin
	}
	return authData, plugin, nil
}

// LoginRequest is the MySQL client login request.
type LoginRequest struct {
	ClientCap    uint16 // 0xA285
	ExtClientCap uint16 // 0x000A
	MaxPacket    uint32 // 0
	Charset      uint8  // 0x2D(45)
	unused       [23]byte
	Username     string
	AuthResp     []byte
	Plugin       string
}

func (lr *LoginRequest) pack() []byte {
	req := bytes.NewBuffer(make([]byte, 0, 128))
	req.Write(convert.LEUint16ToBytes(lr.ClientCap))
	req.Write(convert.LEUint16ToBytes(lr.ExtClientCap))
	req.Write(convert.LEUint32ToBytes(lr.MaxPacket))
	req.WriteByte(lr.Charset)
	req.Write(lr.unused[:])
	req.WriteString(lr.Username)
	req.WriteByte(0x00)
	req.WriteByte(byte(len(lr.AuthResp)))
	req.Write(lr.AuthResp)
	req.WriteString(lr.Plugin)
	req.WriteByte(0x00)
	return req.Bytes()
}

func (mc *mysqlConn) writeLoginRequest(authData []byte, plugin string) error {
	authResp, err := mc.auth(authData, plugin)
	if err != nil {
		return err
	}
	lr := LoginRequest{
		ClientCap:    0xA285,
		ExtClientCap: 0x000A,
		MaxPacket:    maxPacketSize,
		Charset:      0x2D,
		Username:     mc.username,
		AuthResp:     authResp,
		Plugin:       plugin,
	}
	err = mc.writePacket(lr.pack())
	if err != nil {
		return errors.WithMessage(err, "failed to send login request")
	}
	return nil
}

func (mc *mysqlConn) handleAuthResult(oldAuthData []byte, plugin string) error {
	// Read Result Packet
	authData, newPlugin, err := mc.readAuthResult()
	if err != nil {
		return err
	}
	if newPlugin != "" {
		// if CLIENT_PLUGIN_AUTH capability is not supported, no new cipher is
		// sent and we have to keep using the cipher sent in the init packet.
		if authData == nil {
			authData = oldAuthData
		} else {
			// copy data from read buffer to owned slice
			copy(oldAuthData, authData)
		}
		plugin = newPlugin
		authResp, err := mc.auth(authData, plugin)
		if err != nil {
			return err
		}
		err = mc.writePacket(authResp)
		if err != nil {
			return err
		}
		// read result packet
		authData, newPlugin, err = mc.readAuthResult()
		if err != nil {
			return err
		}
		// do not allow to change the auth plugin more than once
		if newPlugin != "" {
			return errors.New("malformed auth result packet")
		}
	}
	switch plugin {
	case "caching_sha2_password":
		return mc.handleCachingSHA2Password(oldAuthData, authData)
	case "sha256_password":
		return mc.handleSHA256Password(oldAuthData, authData)
	default:
		return nil // auth successful
	}
}

func (mc *mysqlConn) readAuthResult() ([]byte, string, error) {
	packet, err := mc.readPacket()
	if err != nil {
		return nil, "", errors.WithMessage(err, "failed to read server greeting")
	}
	if len(packet) < 1 {
		return nil, "", errors.New("malformed authentication result")
	}
	switch packet[0] {
	case iOK:
		return nil, "", nil
	case iAuthMoreData:
		return packet[1:], "", err
	case iEOF:
		// https://dev.mysql.com/doc/internals/en/connection-phase-packets.html
		if len(packet) == 1 {
			return nil, "mysql_old_password", nil
		}
		pluginEndIndex := bytes.IndexByte(packet, 0x00)
		if pluginEndIndex < 0 {
			return nil, "", errors.New("malformed auth result packet")
		}
		plugin := string(packet[1:pluginEndIndex])
		authData := packet[pluginEndIndex+1:]
		return authData, plugin, nil
	default:
		return nil, "", handleErrorPacket(packet)
	}
}

func (mc *mysqlConn) handleCachingSHA2Password(oldAuthData, authData []byte) error {
	switch len(authData) {
	case 0:
		return nil // auth successful
	case 1:
		switch authData[0] {
		case cachingSha2PasswordFastAuthSuccess:
			if err = mc.readResultOK(); err == nil {
				return nil // auth successful
			}
		case cachingSha2PasswordPerformFullAuthentication:
			if mc.cfg.tls != nil || mc.cfg.Net == "unix" {
				// write cleartext auth packet
				err = mc.writeAuthSwitchPacket(append([]byte(mc.cfg.Passwd), 0))
				if err != nil {
					return err
				}
			} else {
				pubKey := mc.cfg.pubKey
				if pubKey == nil {
					// request public key from server
					data, err := mc.buf.takeSmallBuffer(4 + 1)
					if err != nil {
						return err
					}
					data[4] = cachingSha2PasswordRequestPublicKey
					mc.writePacket(data)

					// parse public key
					if data, err = mc.readPacket(); err != nil {
						return err
					}

					block, _ := pem.Decode(data[1:])
					pkix, err := x509.ParsePKIXPublicKey(block.Bytes)
					if err != nil {
						return err
					}
					pubKey = pkix.(*rsa.PublicKey)
				}

				// send encrypted password
				err = mc.sendEncryptedPassword(oldAuthData, pubKey)
				if err != nil {
					return err
				}
			}
			return mc.readResultOK()

		default:
			return ErrMalformPkt
		}
	default:
		return ErrMalformPkt
	}
}

func (mc *mysqlConn) handleSHA256Password(oldAuthData, authData []byte) error {

}

func (mc *mysqlConn) readResultOK() error {
	packet, err := mc.readPacket()
	if err != nil {
		return err
	}
	if packet[0] == iOK {
		return nil
	}
	return handleErrorPacket(packet)
}
