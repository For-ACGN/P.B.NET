package guid

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"os"
	"sync"
	"time"

	"project/internal/convert"
	"project/internal/random"
	"project/internal/xpanic"
)

// +------------+-------------+----------+------------------+------------+
// | hash(part) | PID(hashed) |  random  | timestamp(int64) | ID(uint64) |
// +------------+-------------+----------+------------------+------------+
// |  8 bytes   |   4 bytes   |  8 bytes |      8 bytes     |  4 bytes   |
// +------------+-------------+----------+------------------+------------+

// Size is the generated GUID size, it is not standard.
const Size int = 8 + 4 + 8 + 8 + 4

var zeroGUID = GUID{}

// GUID is the generated GUID, it is not standard size.
type GUID [Size]byte

// Write is used to copy []byte to guid.
func (guid *GUID) Write(b []byte) error {
	if len(b) != Size {
		return errors.New("invalid byte slice size")
	}
	copy(guid[:], b)
	return nil
}

// Print is used to print GUID with prefix.
// Output:
// GUID: BF0AF7928C30AA6B1027DE8D6789F09202262591000000005E6C65F8002AD680
func (guid *GUID) Print() string {
	// 6 = len("GUID: ")
	dst := make([]byte, Size*2+6)
	copy(dst, "GUID: ")
	hex.Encode(dst[6:], guid[:])
	return string(bytes.ToUpper(dst))
}

// Hex is used to encode GUID to a hex string.
// Output:
// BF0AF7928C30AA6B1027DE8D6789F09202262591000000005E6C65F8002AD680
func (guid *GUID) Hex() string {
	dst := make([]byte, Size*2) // add a "\n"
	hex.Encode(dst, guid[:])
	return string(bytes.ToUpper(dst))
}

// Timestamp is used to get timestamp in the GUID.
func (guid *GUID) Timestamp() int64 {
	return int64(binary.BigEndian.Uint64(guid[20:28]))
}

// IsZero is used to check this guid is [0, 0, ...., 0].
func (guid *GUID) IsZero() bool {
	return *guid == zeroGUID
}

// MarshalJSON is used to implement JSON Marshaler interface.
func (guid GUID) MarshalJSON() ([]byte, error) {
	const quotation = 34 // ASCII
	dst := make([]byte, 2*Size+2)
	dst[0] = quotation
	hex.Encode(dst[1:], guid[:])
	dst[2*Size+1] = quotation
	return bytes.ToUpper(dst), nil
}

// UnmarshalJSON is used to implement JSON Unmarshaler interface.
func (guid *GUID) UnmarshalJSON(data []byte) error {
	if len(data) != 2*Size+2 {
		return errors.New("invalid size about guid")
	}
	_, err := hex.Decode(guid[:], data[1:2*Size+1])
	return err
}

// Generator is the GUID generator.
type Generator struct {
	now    func() time.Time
	nowRWM sync.RWMutex
	rand   *random.Rand
	guidCh chan *GUID

	// calculate by NewGenerator
	head []byte

	// id initialize random and self-add
	// when generate but not continuous
	id uint32

	// guid cache pool
	cachePool sync.Pool

	stopSignal chan struct{}
	closeOnce  sync.Once
	wg         sync.WaitGroup
}

// NewGenerator is used to create a GUID generator.
// size is the guid channel buffer size, now is used
// to get timestamp, if now is nil, use time.Now.
func NewGenerator(size int, now func() time.Time) *Generator {
	g := Generator{
		rand:       random.NewRand(),
		stopSignal: make(chan struct{}),
	}
	if now != nil {
		g.now = now
	} else {
		g.now = time.Now
	}
	if size < 1 {
		g.guidCh = make(chan *GUID, 1)
	} else {
		g.guidCh = make(chan *GUID, size)
	}
	// calculate head (8+4 PID)
	hash := sha256.New()
	for i := 0; i < 4096; i++ {
		hash.Write(g.rand.Bytes(64))
	}
	g.head = make([]byte, 0, 8)
	g.head = append(g.head, hash.Sum(nil)[:8]...)
	hash.Write(convert.BEInt64ToBytes(int64(os.Getpid())))
	g.head = append(g.head, hash.Sum(nil)[:4]...)
	// <security> initialize random ID for prevent leak some
	// information like Node, Beacon and Controller Boot time.
	hash.Reset()
	for i := 0; i < 16; i++ {
		hash.Write(g.rand.Bytes(16))
	}
	g.id = convert.BEBytesToUint32(hash.Sum(nil)[:4])
	// initialize guid cache pool
	g.cachePool.New = func() interface{} {
		return new(GUID)
	}
	// start generating
	g.wg.Add(1)
	go g.generateLoop()
	return &g
}

func (g *Generator) generateLoop() {
	defer g.wg.Done()
	defer func() {
		if r := recover(); r != nil {
			xpanic.Log(r, "Generator.generateLoop")
			// restart
			time.Sleep(time.Second)
			g.wg.Add(1)
			go g.generateLoop()
		}
	}()
	for {
		g.id += uint32(g.rand.Int(1024))
		guid := g.cachePool.Get().(*GUID)
		copy(guid[:], g.head)
		copy(guid[12:20], g.rand.Bytes(8))
		// reserve timestamp guid[20:28]
		binary.BigEndian.PutUint32(guid[28:32], g.id)
		select {
		case g.guidCh <- guid:
		case <-g.stopSignal:
			return
		}
	}
}

// Get is used to get a GUID, if guid generator closed, it will return zero guid.
func (g *Generator) Get() *GUID {
	guid := <-g.guidCh
	if guid == nil {
		return new(GUID)
	}
	g.nowRWM.RLock()
	defer g.nowRWM.RUnlock()
	if g.now == nil {
		return new(GUID)
	}
	binary.BigEndian.PutUint64(guid[20:28], uint64(g.now().Unix()))
	return guid
}

// Put is used to put useless GUID to cache pool.
func (g *Generator) Put(guid *GUID) {
	g.cachePool.Put(guid)
}

// Close is used to close guid generator.
func (g *Generator) Close() {
	g.closeOnce.Do(func() {
		close(g.stopSignal)
		g.wg.Wait()
		close(g.guidCh)
		g.nowRWM.Lock()
		defer g.nowRWM.Unlock()
		g.now = nil
	})
}
