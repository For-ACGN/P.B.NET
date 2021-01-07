package rand

import (
	cr "crypto/rand"
	"io"
	"math/rand"
	"time"
)

// Reader is a global, shared instance of a cryptographically
// secure random number generator, it will read 4 * len(b) and
// select random byte but length is equal with len(b).
var Reader io.Reader = reader{}

type reader struct{}

func (reader) Read(b []byte) (int, error) {
	l := len(b)
	size := 4 * l
	buf := make([]byte, size)
	_, err := io.ReadFull(cr.Reader, buf)
	if err != nil {
		return 0, err
	}
	rd := rand.New(rand.NewSource(time.Now().UnixNano())) // #nosec
	for i := 0; i < l; i++ {
		b[i] = buf[rd.Intn(size)]
	}
	return l, nil
}

// Read is a helper function that calls Reader.Read using io.ReadFull.
// On return, n == len(b) if and only if err == nil.
func Read(b []byte) (n int, err error) {
	return io.ReadFull(Reader, b)
}
