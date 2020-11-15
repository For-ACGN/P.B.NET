package security

import (
	"reflect"
	"runtime"
	"sync"
	"unsafe"

	"project/internal/random"
)

var memory *Memory

func init() {
	memory = NewMemory()
	PaddingMemory()
	FlushMemory()
}

// Memory is used to padding memory for randomized memory address.
type Memory struct {
	rand    *random.Rand
	padding map[string][]byte
	mu      sync.Mutex
}

// NewMemory is used to create Memory.
func NewMemory() *Memory {
	mem := &Memory{
		rand:    random.NewRand(),
		padding: make(map[string][]byte),
	}
	mem.Padding()
	return mem
}

// Padding is used to padding memory.
func (m *Memory) Padding() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := 0; i < 16; i++ {
		data := m.rand.Bytes(8 + m.rand.Int(256))
		m.padding[m.rand.String(8)] = data
	}
}

// Flush is used to flush memory.
func (m *Memory) Flush() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.padding = make(map[string][]byte)
}

// PaddingMemory is used to alloc memory.
func PaddingMemory() {
	memory.Padding()
}

// FlushMemory is used to flush global memory.
func FlushMemory() {
	memory.Flush()
}

// CoverBytes is used to cover byte slice if byte slice has secret.
func CoverBytes(b []byte) {
	for i := 0; i < len(b); i++ {
		b[i] = 0
	}
}

// CoverString is used to cover string if string has secret.
// Don't cover string about map key, or maybe trigger data race.
func CoverString(str string) {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&str)) // #nosec
	var bs []byte
	bsh := (*reflect.SliceHeader)(unsafe.Pointer(&bs)) // #nosec
	bsh.Data = sh.Data
	bsh.Len = sh.Len
	bsh.Cap = sh.Len
	CoverBytes(bs)
	runtime.KeepAlive(&str)
}

// CoverRunes is used to cover []rune if it has secret.
func CoverRunes(r []rune) {
	for i := 0; i < len(r); i++ {
		r[i] = 0
	}
}

// Bytes make byte slice discontinuous, it safe for use by multiple goroutines.
type Bytes struct {
	data  map[int]byte
	len   int
	cache sync.Pool
}

// NewBytes is used to create security Bytes.
func NewBytes(b []byte) *Bytes {
	l := len(b)
	bytes := Bytes{
		data: make(map[int]byte, l),
		len:  l,
	}
	for i := 0; i < l; i++ {
		bytes.data[i] = b[i]
	}
	bytes.cache.New = func() interface{} {
		b := make([]byte, l)
		return &b
	}
	return &bytes
}

// Get is used to get stored byte slice.
func (b *Bytes) Get() []byte {
	bytes := *b.cache.Get().(*[]byte)
	for i := 0; i < b.len; i++ {
		bytes[i] = b.data[i]
	}
	return bytes
}

// Put is used to put byte slice to cache, slice will be covered.
func (b *Bytes) Put(bytes []byte) {
	for i := 0; i < b.len; i++ {
		bytes[i] = 0
	}
	b.cache.Put(&bytes)
}

// Len is used to get the bytes length.
func (b *Bytes) Len() int {
	return b.len
}

func (b *Bytes) String() string {
	bytes := b.Get()
	defer b.Put(bytes)
	return string(bytes)
}
