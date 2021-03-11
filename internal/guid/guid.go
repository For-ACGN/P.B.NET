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
// | hash(part) | PID(hashed) |  random  | timestamp(int64) | ID(uint32) |
// +------------+-------------+----------+------------------+------------+
// |  8 bytes   |   4 bytes   |  8 bytes |      8 bytes     |  4 bytes   |
// +------------+-------------+----------+------------------+------------+

// Size is the generated guid size, it is not standard.
const Size = 8 + 4 + 8 + 8 + 4

// GUID is the generated guid, it is not standard size.
type GUID [Size]byte

// Write is used to copy []byte to guid.
func (guid *GUID) Write(b []byte) error {
	if len(b) != Size {
		return errors.New("byte slice size is not equal of guid size")
	}
	copy(guid[:], b)
	return nil
}

// String is used to encode guid to a hex string.
// Output:
// BF0AF7928C30AA6B1027DE8D6789F09202262591000000005E6C65F8002AD680
func (guid *GUID) String() string {
	dst := make([]byte, Size*2)
	hex.Encode(dst, guid[:])
	return string(bytes.ToUpper(dst))
}

// Print is used to print guid with prefix.
// Output:
// GUID: BF0AF7928C30AA6B1027DE8D6789F09202262591000000005E6C65F8002AD680
func (guid *GUID) Print() string {
	dst := make([]byte, Size*2+6) // 6 = len("GUID: ")
	copy(dst, "GUID: ")
	hex.Encode(dst[6:], guid[:])
	return string(bytes.ToUpper(dst))
}

// Timestamp is used to get timestamp in the guid.
func (guid *GUID) Timestamp() int64 {
	return int64(binary.BigEndian.Uint64(guid[20:28]))
}

// zeroGUID is the zero guid for improve GUID.IsZero() performance.
var zeroGUID = GUID{}

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
		return errors.New("invalid guid size")
	}
	_, err := hex.Decode(guid[:], data[1:2*Size+1])
	return err
}

// Generator is the guid generator.
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
	stopOnce   sync.Once
	wg         sync.WaitGroup
}

// NewGenerator is used to create a guid generator.
// size is the guid channel buffer size, now is used
// to get timestamp, if now is nil, use time.Now.
func NewGenerator(size int, now func() time.Time) *Generator {
	gen := Generator{
		rand:       random.NewRand(),
		stopSignal: make(chan struct{}),
	}
	if now != nil {
		gen.now = now
	} else {
		gen.now = time.Now
	}
	if size < 1 {
		gen.guidCh = make(chan *GUID, 1)
	} else {
		gen.guidCh = make(chan *GUID, size)
	}
	// calculate head (8+4 PID)
	hash := sha256.New()
	for i := 0; i < 4096; i++ {
		hash.Write(gen.rand.Bytes(64))
	}
	gen.head = make([]byte, 0, 8)
	gen.head = append(gen.head, hash.Sum(nil)[:8]...)
	hash.Write(convert.BEInt64ToBytes(int64(os.Getpid())))
	gen.head = append(gen.head, hash.Sum(nil)[:4]...)
	// <security> initialize random ID for prevent leak some
	// information like Node, Beacon and Controller Boot time.
	hash.Reset()
	for i := 0; i < 16; i++ {
		hash.Write(gen.rand.Bytes(16))
	}
	id := hash.Sum(nil)[:convert.Uint32Size]
	gen.id = convert.BEBytesToUint32(id)
	// initialize guid cache pool
	gen.cachePool.New = func() interface{} {
		return new(GUID)
	}
	// start generating
	gen.wg.Add(1)
	go gen.generateLoop()
	return &gen
}

func (gen *Generator) generateLoop() {
	defer gen.wg.Done()
	defer func() {
		if r := recover(); r != nil {
			xpanic.Log(r, "Generator.generateLoop")
			// restart
			time.Sleep(time.Second)
			gen.wg.Add(1)
			go gen.generateLoop()
		}
	}()
	for {
		gen.id += gen.rand.Uint32()
		guid := gen.cachePool.Get().(*GUID)
		copy(guid[:], gen.head)
		copy(guid[12:20], gen.rand.Bytes(8))
		// reserve timestamp guid[20:28]
		binary.BigEndian.PutUint32(guid[28:32], gen.id)
		select {
		case gen.guidCh <- guid:
		case <-gen.stopSignal:
			return
		}
	}
}

// Get is used to get a guid, if generator stopped, it will return random guid.
func (gen *Generator) Get() *GUID {
	guid := <-gen.guidCh
	if guid == nil {
		guid = new(GUID)
		copy(guid[:], gen.rand.Bytes(Size))
		return guid
	}
	gen.nowRWM.RLock()
	defer gen.nowRWM.RUnlock()
	var unix int64
	if gen.now != nil {
		unix = gen.now().Unix()
	} else {
		unix = time.Now().Unix()
	}
	binary.BigEndian.PutUint64(guid[20:28], uint64(unix))
	return guid
}

// Put is used to put useless guid to cache pool, it is not necessary.
func (gen *Generator) Put(guid *GUID) {
	gen.cachePool.Put(guid)
}

// Stop is used to stop guid generator.
func (gen *Generator) Stop() {
	gen.stopOnce.Do(func() {
		close(gen.stopSignal)
		gen.wg.Wait()
		close(gen.guidCh)
		gen.nowRWM.Lock()
		defer gen.nowRWM.Unlock()
		gen.now = nil
	})
}
