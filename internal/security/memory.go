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
// If it is a constant string, it will panic because write invalid memory.
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

// NewBytes is used to create security byte slice.
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

// Get is used to get stored byte slice, remember call put after use.
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

// String make string discontinuous, it safe for use by multiple goroutines.
type String struct {
	data  map[int]byte
	len   int
	cache sync.Pool
}

// NewString is used to create security string.
func NewString(s string) *String {
	l := len(s)
	str := String{
		data: make(map[int]byte, l),
		len:  l,
	}
	for i := 0; i < l; i++ {
		str.data[i] = s[i]
	}
	str.cache.New = func() interface{} {
		b := make([]byte, l)
		return &b
	}
	return &str
}

// Get is used to get stored string, remember to call put after use.
func (s *String) Get() string {
	b := *s.cache.Get().(*[]byte)
	defer s.cache.Put(&b)
	for i := 0; i < s.len; i++ {
		b[i] = s.data[i]
	}
	str := string(b)
	// cover byte slice at once
	for i := 0; i < s.len; i++ {
		b[i] = 0
	}
	return str
}

// Put is used to cover string, it is a shortcut.
func (s *String) Put(str string) {
	CoverString(str)
}

// GetBytes is used to get string and return byte slice.
func (s *String) GetBytes() []byte {
	b := *s.cache.Get().(*[]byte)
	for i := 0; i < s.len; i++ {
		b[i] = s.data[i]
	}
	return b
}

// PutBytes is used to put byte slice that get from GetByte.
func (s *String) PutBytes(b []byte) {
	for i := 0; i < s.len; i++ {
		b[i] = 0
	}
	s.cache.Put(&b)
}

// Len is used to get the string length.
func (s *String) Len() int {
	return s.len
}
