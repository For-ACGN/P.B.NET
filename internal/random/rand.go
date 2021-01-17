package random

import (
	cr "crypto/rand"
	"crypto/sha256"
	"io"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"project/internal/convert"
	"project/internal/xpanic"
)

// Rand is used to generate random data. It is multi goroutine safe.
type Rand struct {
	rand *rand.Rand
	mu   sync.Mutex
}

// NewRand is used to create a Rand.
func NewRand() *Rand {
	const (
		goroutines = 4
		times      = 128
	)
	data := make(chan []byte, 16)
	for i := 0; i < goroutines; i++ {
		go sendData(data, times)
	}
	timer := time.NewTimer(time.Second)
	defer timer.Stop()
	hash := sha256.New()
read:
	for i := 0; i < goroutines*times; i++ {
		timer.Reset(time.Second)
		select {
		case d := <-data:
			if d != nil {
				hash.Write(d)
			}
		case <-timer.C:
			break read
		}
	}
	n, _ := io.CopyN(hash, cr.Reader, 512)
	hash.Write([]byte{byte(n)})
	hashData := hash.Sum(nil)
	r := rand.New(rand.NewSource(time.Now().UnixNano())) // #nosec
	selected := make([]byte, convert.Int64Size)
	for i := 0; i < convert.Int64Size; i++ {
		selected[i] = hashData[r.Intn(sha256.Size)]
	}
	seed := convert.BEBytesToInt64(selected)
	return &Rand{rand: rand.New(rand.NewSource(seed))} // #nosec
}

func sendData(data chan<- []byte, times int) {
	defer func() {
		if r := recover(); r != nil {
			xpanic.Log(r, "sendData")
		}
	}()
	timer := time.NewTimer(time.Second)
	defer timer.Stop()
	r := rand.New(rand.NewSource(time.Now().UnixNano())) // #nosec
	count := 0
	for i := 0; i < times; i++ {
		timer.Reset(time.Second)
		select {
		case data <- []byte{byte(r.Intn(256) + i)}:
		case <-timer.C:
			return
		}
		// schedule manually
		if count > 16 {
			runtime.Gosched()
			count = 0
		} else {
			count++
		}
	}
}

// Bytes is used to generate random byte slice that size = n.
func (r *Rand) Bytes(n int) []byte {
	if n < 1 {
		return nil
	}
	result := make([]byte, n)
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := 0; i < n; i++ {
		ri := r.rand.Intn(256)
		result[i] = byte(ri)
	}
	return result
}

// String returns a string that only include 0-9, A-Z and a-z.
func (r *Rand) String(n int) string {
	if n < 1 {
		return ""
	}
	result := make([]rune, n)
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := 0; i < n; i++ {
		// after space
		ri := 33 + r.rand.Intn(90)
		switch {
		case ri >= '0' && ri <= '9':
		case ri >= 'A' && ri <= 'Z':
		case ri >= 'a' && ri <= 'z':
		default:
			i--
			continue
		}
		result[i] = rune(ri)
	}
	return string(result)
}

// Intn returns, as an int, a non-negative pseudo-random number in [0,n).
func (r *Rand) Intn(n int) int {
	if n < 1 {
		return 0
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.rand.Intn(n)
}

// Int31n returns, as an int32, a non-negative pseudo-random number in [0,n).
func (r *Rand) Int31n(n int32) int32 {
	if n < 1 {
		return 0
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.rand.Int31n(n)
}

// Int63n returns, as an int64, a non-negative pseudo-random number in [0,n).
func (r *Rand) Int63n(n int64) int64 {
	if n < 1 {
		return 0
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.rand.Int63n(n)
}

// Int returns a non-negative pseudo-random int.
func (r *Rand) Int() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.rand.Int()
}

// Int31 returns a non-negative pseudo-random 31-bit integer as an int32.
func (r *Rand) Int31() int32 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.rand.Int31()
}

// Int63 returns a non-negative pseudo-random 63-bit integer as an int64.
func (r *Rand) Int63() int64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.rand.Int63()
}

// Uint32 returns a pseudo-random 32-bit value as a uint32.
func (r *Rand) Uint32() uint32 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.rand.Uint32()
}

// Uint64 returns a pseudo-random 64-bit value as a uint64.
func (r *Rand) Uint64() uint64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.rand.Uint64()
}

// Float32 returns, as a float32, a pseudo-random number in [0.0,1.0).
func (r *Rand) Float32() float32 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.rand.Float32()
}

// Float64 returns, as a float64, a pseudo-random number in [0.0,1.0).
func (r *Rand) Float64() float64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.rand.Float64()
}

// NormFloat64 returns a normally distributed float64 in
// the range -math.MaxFloat64 through +math.MaxFloat64 inclusive,
// with standard normal distribution (mean = 0, stddev = 1).
// To produce a different normal distribution, callers can
// adjust the output using:
//
//  sample = NormFloat64() * desiredStdDev + desiredMean
//
func (r *Rand) NormFloat64() float64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.rand.NormFloat64()
}

// ExpFloat64 returns an exponentially distributed float64 in the range
// (0, +math.MaxFloat64] with an exponential distribution whose rate parameter
// (lambda) is 1 and whose mean is 1/lambda (1).
// To produce a distribution with a different rate parameter,
// callers can adjust the output using:
//
//  sample = ExpFloat64() / desiredRateParameter
//
func (r *Rand) ExpFloat64() float64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.rand.ExpFloat64()
}

// Perm returns, as a slice of n int, a pseudo-random permutation of the integers [0,n).
func (r *Rand) Perm(n int) []int {
	if n < 1 {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.rand.Perm(n)
}

// Shuffle pseudo-randomizes the order of elements.
// n is the number of elements. Shuffle panics if n < 0.
// swap swaps the elements with indexes i and j.
func (r *Rand) Shuffle(n int, swap func(i, j int)) {
	if n < 1 {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rand.Shuffle(n, swap)
}

var gRand = NewRand()

// Bytes is used to generate random []byte that size = n.
func Bytes(n int) []byte {
	return gRand.Bytes(n)
}

// String returns a string that only include 0-9, A-Z and a-z.
func String(n int) string {
	return gRand.String(n)
}

// Intn returns, as an int, a non-negative pseudo-random number in [0,n).
func Intn(n int) int {
	return gRand.Intn(n)
}

// Int31n returns, as an int32, a non-negative pseudo-random number in [0,n).
func Int31n(n int32) int32 {
	return gRand.Int31n(n)
}

// Int63n returns, as an int64, a non-negative pseudo-random number in [0,n).
func Int63n(n int64) int64 {
	return gRand.Int63n(n)
}

// Int returns a non-negative pseudo-random int.

// Int31 returns a non-negative pseudo-random 31-bit integer as an int32.

// Int63 returns a non-negative pseudo-random 63-bit integer as an int64.
func Int63() int64 {
	return gRand.Int63()
}

// Uint32 returns a pseudo-random 32-bit value as a uint32.

// Uint64 returns a pseudo-random 64-bit value as a uint64.
func Uint64() uint64 {
	return gRand.Uint64()
}
