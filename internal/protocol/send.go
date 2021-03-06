package protocol

import (
	"bytes"
	"crypto/sha256"
	"errors"

	"project/internal/convert"
	"project/internal/crypto/aes"
	"project/internal/guid"
)

const (
	// SendMinBufferSize is the sender and worker minimum buffer size.
	SendMinBufferSize = 2*guid.Size + flagSize + sha256.Size + aes.BlockSize

	// AcknowledgeSize is the acknowledge packet size.
	AcknowledgeSize = 3*guid.Size + sha256.Size

	// IndexSize is len(uint64).
	IndexSize = 8

	// QuerySize is the query packet size.
	QuerySize = 2*guid.Size + IndexSize + sha256.Size
)

// ---------------------------interactive mode---------------------------------
// In interactive mode, role send message will use "Send" to send message
// and wait "Acknowledge" to make sure message has reached.
// Node is always in interactive mode.
// In default, Beacon use query mode, you can switch interactive mode manually.

// +----------+-----------+---------+-------------+---------+
// |   GUID   | role GUID | deflate | HMAC-SHA256 | message |
// +----------+-----------+---------+-------------+---------+
// | 32 bytes | 32 bytes  |  byte   |  32 bytes   |   var   |
// +----------+-----------+---------+-------------+---------+

// Send is used to send messages in interactive mode.
//
// When Controller use it, RoleGUID = receiver role GUID.
// Controller encrypt Message with Node or Beacon session key.
// When Node use it, RoleGUID = its GUID.
// Node encrypt Message with it's session key.
// When Beacon use it, RoleGUID = its GUID.
// Beacon encrypt Message with it's session key.
type Send struct {
	GUID     guid.GUID // prevent duplicate handle it
	RoleGUID guid.GUID // receiver GUID
	Deflate  byte      // use deflate to compress it(0=disable, 1=enable)
	Hash     []byte    // hash(GUID + RoleGUID + Deflate + Message)
	Message  []byte    // use AES to encrypt it(maybe compressed first)
}

// NewSend is used to create a send, Unpack() need it,
// if only used to Pack(), use new(Send).
func NewSend() *Send {
	return &Send{
		Hash:    make([]byte, sha256.Size),
		Message: make([]byte, 2*aes.BlockSize),
	}
}

// Pack is used to pack Send to *bytes.Buffer.
func (s *Send) Pack(buf *bytes.Buffer) {
	buf.Write(s.GUID[:])
	buf.Write(s.RoleGUID[:])
	buf.WriteByte(s.Deflate)
	buf.Write(s.Hash)
	buf.Write(s.Message)
}

// Unpack is used to unpack []byte to Send.
func (s *Send) Unpack(data []byte) error {
	if len(data) < 2*guid.Size+flagSize+sha256.Size+aes.BlockSize {
		return errors.New("invalid send packet size")
	}
	copy(s.GUID[:], data[:guid.Size])
	copy(s.RoleGUID[:], data[guid.Size:2*guid.Size])
	s.Deflate = data[2*guid.Size]
	copy(s.Hash, data[2*guid.Size+flagSize:2*guid.Size+flagSize+sha256.Size])
	message := data[2*guid.Size+flagSize+sha256.Size:]
	mLen := len(message)
	smLen := len(s.Message)
	if cap(s.Message) >= mLen {
		switch {
		case smLen > mLen:
			copy(s.Message, message)
			s.Message = s.Message[:mLen]
		case smLen == mLen:
			copy(s.Message, message)
		case smLen < mLen:
			s.Message = append(s.Message[:0], message...)
		}
	} else {
		s.Message = make([]byte, mLen)
		copy(s.Message, message)
	}
	return nil
}

// Validate is used to validate send fields.
func (s *Send) Validate() error {
	if s.Deflate > 1 {
		return errors.New("invalid deflate flag")
	}
	if len(s.Hash) != sha256.Size {
		return errors.New("invalid hmac hash size")
	}
	l := len(s.Message)
	if l < aes.BlockSize || l%aes.BlockSize != 0 {
		return errors.New("invalid message size")
	}
	return nil
}

// SendResponse is use to get send response.
type SendResponse struct {
	Role Role
	GUID *guid.GUID // Role GUID
	Err  error
}

// SendResult is use to get send result, it include all SendResponse.
type SendResult struct {
	Success   int
	Responses []*SendResponse
	Err       error
}

// Clean is used to clean SendResult for sync.Pool.
func (sr *SendResult) Clean() {
	sr.Success = 0
	sr.Responses = nil
	sr.Err = nil
}

// +----------+-----------+-----------+-------------+
// |   GUID   | role GUID | send GUID | HMAC-SHA256 |
// +----------+-----------+-----------+-------------+
// | 32 bytes | 32 bytes  | 32 bytes  |  32 bytes   |
// +----------+-----------+-----------+-------------+

// Acknowledge is used to acknowledge sender that receiver has receive this message.
//
// When Controller use it, RoleGUID = sender GUID.
// When Node use it, RoleGUID = it's GUID.
// When Beacon use it, RoleGUID = it's GUID.
type Acknowledge struct {
	GUID     guid.GUID // prevent duplicate handle it
	RoleGUID guid.GUID // sender GUID
	SendGUID guid.GUID // structure Send.GUID
	Hash     []byte    // hash(GUID + RoleGUID + SendGUID)
}

// NewAcknowledge is used to create a acknowledge, Unpack() need it,
// if only used to Pack(), use new(Acknowledge).
func NewAcknowledge() *Acknowledge {
	return &Acknowledge{
		Hash: make([]byte, sha256.Size),
	}
}

// Pack is used to pack Acknowledge to *bytes.Buffer.
func (ack *Acknowledge) Pack(buf *bytes.Buffer) {
	buf.Write(ack.GUID[:])
	buf.Write(ack.RoleGUID[:])
	buf.Write(ack.SendGUID[:])
	buf.Write(ack.Hash)
}

// Unpack is used to unpack []byte to Acknowledge.
func (ack *Acknowledge) Unpack(data []byte) error {
	if len(data) != AcknowledgeSize {
		return errors.New("invalid acknowledge packet size")
	}
	copy(ack.GUID[:], data[:guid.Size])
	copy(ack.RoleGUID[:], data[guid.Size:2*guid.Size])
	copy(ack.SendGUID[:], data[2*guid.Size:3*guid.Size])
	copy(ack.Hash, data[3*guid.Size:3*guid.Size+sha256.Size])
	return nil
}

// Validate is used to validate acknowledge fields.
func (ack *Acknowledge) Validate() error {
	if len(ack.Hash) != sha256.Size {
		return errors.New("invalid hmac hash size")
	}
	return nil
}

// AcknowledgeResponse is use to get acknowledge response.
type AcknowledgeResponse struct {
	Role Role
	GUID *guid.GUID // Role GUID
	Err  error
}

// AcknowledgeResult is use to get acknowledge result, it include all AcknowledgeResponse.
type AcknowledgeResult struct {
	Success   int
	Responses []*AcknowledgeResponse
	Err       error
}

// Clean is used to clean AcknowledgeResult for sync.Pool.
func (ar *AcknowledgeResult) Clean() {
	ar.Success = 0
	ar.Responses = nil
	ar.Err = nil
}

// --------------------------------query mode----------------------------------
// In query mode, beacon send "Query" to Nodes, Nodes will forward all Nodes.
// if Controller is online(Controller connect Nodes), Controller will answer
// to all Nodes, then Node that Beacon connected will return "Answer" to Beacon.
// In default, Beacon use query mode, you can switch interactive mode manually.
// Only Beacon will use it, because Node always in interactive mode.

// +----------+-------------+---------+-------------+
// |   GUID   | Beacon GUID |  index  | HMAC-SHA256 |
// +----------+-------------+---------+-------------+
// | 32 bytes |  32 bytes   | 8 bytes |  32 bytes   |
// +----------+-------------+---------+-------------+

// Query is used to query message from controller.
type Query struct {
	GUID       guid.GUID // prevent duplicate handle it
	BeaconGUID guid.GUID // beacon GUID
	Index      uint64    // controller will delete message < this index
	Hash       []byte    // sign(GUID + BeaconGUID + Index)
}

// NewQuery is used to create a query, Unpack() need it,
// if only used to Pack(), use new(Query).
func NewQuery() *Query {
	return &Query{
		Hash: make([]byte, sha256.Size),
	}
}

// Pack is used to pack Query to *bytes.Buffer.
func (q *Query) Pack(buf *bytes.Buffer) {
	buf.Write(q.GUID[:])
	buf.Write(q.BeaconGUID[:])
	buf.Write(convert.BEUint64ToBytes(q.Index))
	buf.Write(q.Hash)
}

// Unpack is used to unpack []byte to Query.
func (q *Query) Unpack(data []byte) error {
	if len(data) != QuerySize {
		return errors.New("invalid query packet size")
	}
	copy(q.GUID[:], data[:guid.Size])
	copy(q.BeaconGUID[:], data[guid.Size:2*guid.Size])
	q.Index = convert.BEBytesToUint64(data[2*guid.Size : 2*guid.Size+IndexSize])
	copy(q.Hash, data[2*guid.Size+IndexSize:2*guid.Size+IndexSize+sha256.Size])
	return nil
}

// Validate is used to validate query fields.
func (q *Query) Validate() error {
	if len(q.Hash) != sha256.Size {
		return errors.New("invalid hmac hash size")
	}
	return nil
}

// QueryResponse is use to get query response.
type QueryResponse struct {
	Role Role
	GUID *guid.GUID // Role GUID
	Err  error
}

// QueryResult is use to get query result, it include all QueryResponse.
type QueryResult struct {
	Success   int
	Responses []*QueryResponse
	Err       error
}

// Clean is used to clean QueryResult for sync.Pool.
func (qr *QueryResult) Clean() {
	qr.Success = 0
	qr.Responses = nil
	qr.Err = nil
}

// +----------+-------------+---------+---------+-------------+---------+
// |   GUID   | Beacon GUID |  index  | deflate | HMAC-SHA256 | message |
// +----------+-------------+---------+---------+-------------+---------+
// | 32 bytes |  32 bytes   | 8 bytes |  byte   |  32 bytes   |   var   |
// +----------+-------------+---------+---------+-------------+---------+

// Answer is used to return queried message.
type Answer struct {
	GUID       guid.GUID // prevent duplicate handle it
	BeaconGUID guid.GUID // beacon GUID
	Index      uint64    // compare Query.Index
	Deflate    byte      // use deflate to compress it(0=disable, 1=enable)
	Hash       []byte    // sign(GUID + RoleGUID + Index + Deflate + Message)
	Message    []byte    // use AES to encrypt it(maybe compressed first)
}

// NewAnswer is used to create a answer, Unpack() need it,
// if only used to Pack(), use new(Answer).
func NewAnswer() *Answer {
	return &Answer{
		Hash:    make([]byte, sha256.Size),
		Message: make([]byte, 2*aes.BlockSize),
	}
}

// Pack is used to pack Answer to *bytes.Buffer.
func (a *Answer) Pack(buf *bytes.Buffer) {
	buf.Write(a.GUID[:])
	buf.Write(a.BeaconGUID[:])
	buf.Write(convert.BEUint64ToBytes(a.Index))
	buf.WriteByte(a.Deflate)
	buf.Write(a.Hash)
	buf.Write(a.Message)
}

// Unpack is used to unpack []byte to Answer.
func (a *Answer) Unpack(data []byte) error {
	if len(data) < 2*guid.Size+IndexSize+flagSize+sha256.Size+aes.BlockSize {
		return errors.New("invalid answer packet size")
	}
	copy(a.GUID[:], data[:guid.Size])
	copy(a.BeaconGUID[:], data[guid.Size:2*guid.Size])
	a.Index = convert.BEBytesToUint64(data[2*guid.Size : 2*guid.Size+IndexSize])
	a.Deflate = data[2*guid.Size+IndexSize]
	copy(a.Hash, data[2*guid.Size+IndexSize+flagSize:2*guid.Size+IndexSize+flagSize+sha256.Size])
	message := data[2*guid.Size+IndexSize+flagSize+sha256.Size:]
	mLen := len(message)
	amLen := len(a.Message)
	if cap(a.Message) >= mLen {
		switch {
		case amLen > mLen:
			copy(a.Message, message)
			a.Message = a.Message[:mLen]
		case amLen == mLen:
			copy(a.Message, message)
		case amLen < mLen:
			a.Message = append(a.Message[:0], message...)
		}
	} else {
		a.Message = make([]byte, mLen)
		copy(a.Message, message)
	}
	return nil
}

// Validate is used to validate Answer fields.
func (a *Answer) Validate() error {
	if a.Deflate > 1 {
		return errors.New("invalid deflate flag")
	}
	if len(a.Hash) != sha256.Size {
		return errors.New("invalid hmac hash size")
	}
	l := len(a.Message)
	if l < aes.BlockSize || l%aes.BlockSize != 0 {
		return errors.New("invalid message size")
	}
	return nil
}

// AnswerResponse is use to get answer response.
type AnswerResponse struct {
	GUID *guid.GUID // Role GUID
	Err  error
}

// AnswerResult is use to get answer result, it include all AnswerResponse.
type AnswerResult struct {
	Success   int
	Responses []*AnswerResponse
	Err       error
}

// Clean is used to clean AnswerResult for sync.Pool.
func (ar *AnswerResult) Clean() {
	ar.Success = 0
	ar.Responses = nil
	ar.Err = nil
}
