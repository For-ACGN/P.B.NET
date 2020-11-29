package protocol

import (
	"bytes"
	"errors"
	"net"
	"time"

	"project/internal/convert"
)

// +------------+---------+------------+------+
// | frame size | command | [frame id] | data |
// +------------+---------+------------+------+
// |   uint32   |  uint8  |   uint16   |  var |
// +------------+---------+------------+------+
//
// frame size = command + frame id + data
// heartbeat don't need set frame id

// about connection.
const (
	MaxFrameSize = 2 * 1024 * 1024
	SendTimeout  = time.Minute
	RecvTimeout  = 2 * SendTimeout // wait heartbeat send time
	SlotSize     = 1024
	MaxFrameID   = SlotSize - 1

	// don't change
	FrameLenSize    = 4
	FrameCMDSize    = 1
	FrameIDSize     = 2
	FrameHeaderSize = FrameLenSize + FrameCMDSize + FrameIDSize
)

// errors about frame.
var (
	ErrInvalidFrameSize       = errors.New("invalid frame size")
	ErrRecvNullFrame          = errors.New("receive null frame")
	ErrRecvTooLargeFrame      = errors.New("receive too large frame")
	ErrRecvInvalidFrameIDSize = errors.New("receive invalid frame id size")
	ErrRecvInvalidFrameID     = errors.New("receive invalid frame id")
	ErrRecvInvalidReplyID     = errors.New("receive invalid reply id")
	ErrRecvReplyTimeout       = errors.New("receive reply timeout")
	ErrConnClosed             = errors.New("connection closed")
)

// Slot is used to handle frame async
type Slot struct {
	Available chan struct{}
	Reply     chan []byte
	Timer     *time.Timer // receive reply timeout
}

// NewSlots is used to create slot.
func NewSlots() []*Slot {
	slots := make([]*Slot, SlotSize)
	for i := 0; i < SlotSize; i++ {
		slots[i] = &Slot{
			Available: make(chan struct{}, 1),
			Reply:     make(chan []byte, 1),
			Timer:     time.NewTimer(RecvTimeout),
		}
		slots[i].Timer.Stop()
		slots[i].Available <- struct{}{}
	}
	return slots
}

// DestroySlots is used to stop all timers in slots.
func DestroySlots(slots []*Slot) {
	for i := 0; i < SlotSize; i++ {
		slots[i].Timer.Stop()
	}
}

var (
	errNullFrame     = []byte{ConnErrRecvNullFrame}
	errTooLargeFrame = []byte{ConnErrRecvTooLargeFrame}
)

// HandleConn is used to handle frame
func HandleConn(conn net.Conn, handler func([]byte)) {
	const (
		// if data buffer bufSize > this, new buffer
		bufSize    = 4096
		maxBufSize = 4 * bufSize

		// for cmd frameID GUID Hash...
		maxFrameSize = MaxFrameSize + 256

		// client send heartbeat
		heartbeatTimeout = 120 * time.Second
	)
	buffer := make([]byte, bufSize)
	data := bytes.NewBuffer(make([]byte, 0, bufSize))
	var (
		bodySize int
		n        int
		l        int
		err      error
	)
	for {
		_ = conn.SetReadDeadline(time.Now().Add(heartbeatTimeout))
		n, err = conn.Read(buffer)
		if err != nil {
			return
		}
		if n == 0 {
			continue
		}
		data.Write(buffer[:n])
		l = data.Len()
		for {
			if l < FrameLenSize {
				break
			}
			if bodySize == 0 { // avoid duplicate calculations
				bodySize = int(convert.BEBytesToUint32(data.Next(FrameLenSize)))
				if bodySize == 0 {
					handler(errNullFrame)
					return
				}
				if bodySize > maxFrameSize {
					handler(errTooLargeFrame)
					return
				}
			}
			l = data.Len()
			if l < bodySize {
				break
			}
			handler(data.Next(bodySize))
			bodySize = 0
			l = data.Len()
		}
		// flush and write rest
		if bodySize == 0 {
			leftover := data.Bytes()
			if data.Cap() > maxBufSize {
				data = bytes.NewBuffer(make([]byte, 0, bufSize))
			} else {
				data.Reset()
			}
			data.Write(leftover)
		}
	}
}
