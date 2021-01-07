package rand

import (
	cr "crypto/rand"
	"io"
	"math/rand"
	"sync"
	"time"

	"project/internal/random"
)

// Reader is a global, shared instance of a cryptographically
// secure random number generator, it will read bytes and
// select random byte but length is equal with len(b).
var Reader io.Reader = reader{}

var (
	// cachePool is the rand.Rand cache pool.
	cachePool sync.Pool

	// iRand is used to generate seed for rand.New.
	iRand *random.Rand
)

func init() {
	iRand = random.NewRand()
	cachePool.New = func() interface{} {
		seed := iRand.Int64() + time.Now().Unix() + time.Now().UnixNano()
		return rand.New(rand.NewSource(seed)) // #nosec
	}
}

type reader struct{}

func (reader) Read(b []byte) (int, error) {
	_, err := io.ReadFull(cr.Reader, b)
	if err != nil {
		return 0, err
	}
	// get rand.Rand form cache pool
	rd := cachePool.Get().(*rand.Rand)
	defer cachePool.Put(rd)
	// random swap and add random value
	l := len(b)
	rv := rd.Intn(l)
	for i := 0; i < l; i++ {
		b[i], b[l-i-1] = b[rv], b[i]+byte(rv+1024)
		rv++
		if rv >= l {
			rv = 0
		}
	}
	return l, nil
}

// Read is a helper function that calls Reader.Read using io.ReadFull.
// On return, n == len(b) if and only if err == nil.
func Read(b []byte) (n int, err error) {
	return io.ReadFull(Reader, b)
}
