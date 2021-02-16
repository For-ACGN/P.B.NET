package random

import (
	"crypto/sha256"
	"math"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"project/internal/patch/monkey"
)

// copy from internal/testsuite/testsuite.go
func testDeferForPanic(t testing.TB) {
	r := recover()
	require.NotNil(t, r)
	t.Logf("\npanic in %s:\n%s\n", t.Name(), r)
}

func TestNewRand(t *testing.T) {
	t.Run("read timer.C timeout", func(t *testing.T) {
		patch := func(rand.Source) *rand.Rand {
			panic(monkey.Panic)
		}
		pg := monkey.Patch(rand.New, patch)
		defer pg.Unpatch()

		defer testDeferForPanic(t)
		NewRand()
	})

	t.Run("send data timeout", func(t *testing.T) {
		hash := sha256.New()
		patch := func(interface{}, []byte) (int, error) {
			panic(monkey.Panic)
		}
		pg := monkey.PatchInstanceMethod(hash, "Write", patch)
		defer pg.Unpatch()

		defer time.Sleep(3 * time.Second)
		defer testDeferForPanic(t)
		NewRand()
	})
}

func TestRand_Bytes(t *testing.T) {
	bytes := Bytes(10)
	t.Log(bytes)
	require.Len(t, bytes, 10)

	bytes = Bytes(-1)
	require.Len(t, bytes, 0)
}

func TestRand_String(t *testing.T) {
	str := String(32768)
	require.Len(t, str, 32768)

	str = String(-1)
	require.Len(t, str, 0)
}

func TestRand_Password(t *testing.T) {
	pwd := Password(4096)
	require.Len(t, pwd, 4096)

	pwd = Password(-1)
	require.Len(t, pwd, 12)
}

func TestRand_Bool(t *testing.T) {
	m := make(map[bool]bool, 2)
	for i := 0; i < 1000; i++ {
		m[Bool()] = true
	}

	require.True(t, m[true])
	require.True(t, m[false])
}

func TestRand_Intn(t *testing.T) {
	for i := 0; i < 1000; i++ {
		v := Intn(1<<63 - 1)
		require.True(t, v >= 0 && v < 1<<63-1)

		v = Intn(1<<33 - 1)
		require.True(t, v >= 0 && v < 1<<33-1)
	}

	require.True(t, Intn(-1) == 0)

	m := make(map[int]bool, 10)
	for i := 0; i < 10000; i++ {
		m[Intn(10)] = true
	}
	for i := 0; i < 10; i++ {
		require.True(t, m[i])
	}
}

func TestRand_Int7n(t *testing.T) {
	for i := 0; i < 1000; i++ {
		v := Int7n(1<<7 - 1)
		require.True(t, v >= 0 && v < 1<<7-1, v)

		v = Int7n(1<<4 - 1)
		require.True(t, v >= 0 && v < 1<<7-1, v)
	}

	require.True(t, Int7n(-1) == 0)

	m := make(map[int8]bool, 10)
	for i := 0; i < 10000; i++ {
		m[Int7n(10)] = true
	}
	for i := int8(0); i < 10; i++ {
		require.True(t, m[i])
	}
}

func TestRand_Int15n(t *testing.T) {
	for i := 0; i < 1000; i++ {
		v := Int15n(1<<15 - 1)
		require.True(t, v >= 0 && v < 1<<15-1, v)

		v = Int15n(1<<8 - 1)
		require.True(t, v >= 0 && v < 1<<8-1, v)
	}

	require.True(t, Int15n(-1) == 0)

	m := make(map[int16]bool, 10)
	for i := 0; i < 10000; i++ {
		m[Int15n(10)] = true
	}
	for i := int16(0); i < 10; i++ {
		require.True(t, m[i])
	}
}

func TestRand_Int31n(t *testing.T) {
	for i := 0; i < 1000; i++ {
		v := Int31n(1<<31 - 1)
		require.True(t, v >= 0 && v < 1<<31-1, v)

		v = Int31n(1<<17 - 1)
		require.True(t, v >= 0 && v < 1<<17-1, v)
	}

	require.True(t, Int31n(-1) == 0)

	m := make(map[int32]bool, 10)
	for i := 0; i < 10000; i++ {
		m[Int31n(10)] = true
	}
	for i := int32(0); i < 10; i++ {
		require.True(t, m[i])
	}
}

func TestRand_Int63n(t *testing.T) {
	for i := 0; i < 1000; i++ {
		v := Int63n(1<<63 - 1)
		require.True(t, v >= 0 && v < 1<<63-1, v)

		v = Int63n(1<<33 - 1)
		require.True(t, v >= 0 && v < 1<<33-1, v)
	}

	require.True(t, Int63n(-1) == 0)

	m := make(map[int64]bool, 10)
	for i := 0; i < 10000; i++ {
		m[Int63n(10)] = true
	}
	for i := int64(0); i < 10; i++ {
		require.True(t, m[i])
	}
}

func TestRand_Uintn(t *testing.T) {
	for i := 0; i < 1000; i++ {
		v := Uintn(1<<63 - 1)
		require.True(t, v < 1<<63-1)

		v = Uintn(1<<33 - 1)
		require.True(t, v < 1<<33-1)

		v = Uintn(1<<64 - 1)
		require.True(t, v < 1<<64-1)
	}

	require.True(t, Uintn(0) == 0)

	m := make(map[uint]bool, 10)
	for i := 0; i < 10000; i++ {
		m[Uintn(10)] = true
	}
	for i := uint(0); i < 10; i++ {
		require.True(t, m[i])
	}
}

func TestRand_Uint8n(t *testing.T) {
	for i := 0; i < 1000; i++ {
		v := Uint8n(1<<8 - 1)
		require.True(t, v < 1<<8-1)

		v = Uint8n(1<<4 - 1)
		require.True(t, v < 1<<4-1)
	}

	require.True(t, Uint8n(0) == 0)

	m := make(map[uint8]bool, 10)
	for i := 0; i < 10000; i++ {
		m[Uint8n(10)] = true
	}
	for i := uint8(0); i < 10; i++ {
		require.True(t, m[i])
	}
}

func TestRand_Uint16n(t *testing.T) {
	for i := 0; i < 1000; i++ {
		v := Uint16n(1<<16 - 1)
		require.True(t, v < 1<<16-1)

		v = Uint16n(1<<8 - 1)
		require.True(t, v < 1<<8-1)
	}

	require.True(t, Uint16n(0) == 0)

	m := make(map[uint16]bool, 10)
	for i := 0; i < 10000; i++ {
		m[Uint16n(10)] = true
	}
	for i := uint16(0); i < 10; i++ {
		require.True(t, m[i])
	}
}

func TestRand_Uint32n(t *testing.T) {
	for i := 0; i < 1000; i++ {
		v := Uint32n(1<<32 - 1)
		require.True(t, v < 1<<32-1)

		v = Uint32n(1<<16 - 1)
		require.True(t, v < 1<<16-1)
	}

	require.True(t, Uint32n(0) == 0)

	m := make(map[uint32]bool, 10)
	for i := 0; i < 10000; i++ {
		m[Uint32n(10)] = true
	}
	for i := uint32(0); i < 10; i++ {
		require.True(t, m[i])
	}
}

func TestRand_Uint64n(t *testing.T) {
	for i := 0; i < 1000; i++ {
		v := Uint64n(1<<63 - 1)
		require.True(t, v < 1<<63-1)

		v = Uint64n(1<<33 - 1)
		require.True(t, v < 1<<33-1)

		v = Uint64n(1<<64 - 1)
		require.True(t, v < 1<<64-1)
	}

	require.True(t, Uint64n(0) == 0)

	m := make(map[uint64]bool, 10)
	for i := uint64(0); i < 10000; i++ {
		m[Uint64n(10)] = true
	}
	for i := uint64(0); i < 10; i++ {
		require.True(t, m[i])
	}
}

func TestRand_Int(t *testing.T) {
	for i := 0; i < 1000; i++ {
		v := Int()
		require.True(t, v >= 0)
	}
}

func TestRand_Int7(t *testing.T) {
	for i := 0; i < 1000; i++ {
		v := Int7()
		require.True(t, v >= 0, v)
	}
}

func TestRand_Int15(t *testing.T) {
	for i := 0; i < 1000; i++ {
		v := Int15()
		require.True(t, v >= 0, v)
	}
}

func TestRand_Int31(t *testing.T) {
	for i := 0; i < 1000; i++ {
		v := Int31()
		require.True(t, v >= 0, v)
	}
}

func TestRand_Int63(t *testing.T) {
	for i := 0; i < 1000; i++ {
		v := Int63()
		require.True(t, v >= 0, v)
	}
}

func TestRand_Uint(t *testing.T) {
	for i := 0; i < 1000; i++ {
		Uint()
	}
}

func TestRand_Unt8(t *testing.T) {
	for i := 0; i < 1000; i++ {
		Uint8()
	}
}

func TestRand_Uint16(t *testing.T) {
	for i := 0; i < 1000; i++ {
		Uint16()
	}
}

func TestRand_Uint32(t *testing.T) {
	for i := 0; i < 1000; i++ {
		Uint32()
	}
}

func TestRand_Uint64(t *testing.T) {
	for i := 0; i < 1000; i++ {
		Uint64()
	}
}

func TestRand_Float32(t *testing.T) {
	for i := 0; i < 1000; i++ {
		v := Float32()
		require.True(t, v >= 0 && v <= 1)
	}
}

func TestRand_Float64(t *testing.T) {
	for i := 0; i < 1000; i++ {
		v := Float64()
		require.True(t, v >= 0 && v <= 1)
	}
}

func TestRand_NormFloat64(t *testing.T) {
	for i := 0; i < 1000; i++ {
		v := NormFloat64()
		require.True(t, v >= -math.MaxFloat64 && v <= math.MaxFloat64)
	}
}

func TestRand_ExpFloat64(t *testing.T) {
	for i := 0; i < 1000; i++ {
		v := ExpFloat64()
		require.True(t, v >= 0 && v <= math.MaxFloat64)
	}
}

func TestRand_Perm(t *testing.T) {
	n := Perm(16)
	for i := 0; i < len(n); i++ {
		require.Less(t, n[i], 16)
	}

	require.Zero(t, Perm(0))
}

func TestRand_Shuffle(t *testing.T) {
	Shuffle(16, func(i, j int) {
		t.Log(i, j)
		require.Less(t, i, 16)
		require.Less(t, j, 16)
	})

	Shuffle(0, nil)
}

func TestRand_Equal(t *testing.T) {
	const n = 64
	result := make(chan int, n)
	for i := 0; i < n; i++ {
		go func() {
			r := NewRand()
			result <- r.Intn(1048576)
		}()
	}
	results := make(map[int]struct{})
	for i := 0; i < n; i++ {
		r := <-result
		_, ok := results[r]
		require.False(t, ok, "appeared value: %d, i: %d", r, i)
		results[r] = struct{}{}
	}
}

func TestRand_Parallel(t *testing.T) {
	r := NewRand()
	wg := sync.WaitGroup{}
	for _, fn := range []func(){
		func() { t.Log(r.Bytes(16)) },
		func() { t.Log(r.String(16)) },
		func() { t.Log(r.Bool()) },
		func() { t.Log(r.Intn(16)) },
		func() { t.Log(r.Int7n(16)) },
		func() { t.Log(r.Int15n(16)) },
		func() { t.Log(r.Int31n(16)) },
		func() { t.Log(r.Int63n(16)) },
		func() { t.Log(r.Uintn(16)) },
		func() { t.Log(r.Uint8n(16)) },
		func() { t.Log(r.Uint16n(16)) },
		func() { t.Log(r.Uint32n(16)) },
		func() { t.Log(r.Uint64n(16)) },
		func() { t.Log(r.Int()) },
		func() { t.Log(r.Int7()) },
		func() { t.Log(r.Int15()) },
		func() { t.Log(r.Int31()) },
		func() { t.Log(r.Int63()) },
		func() { t.Log(r.Uint()) },
		func() { t.Log(r.Uint8()) },
		func() { t.Log(r.Uint16()) },
		func() { t.Log(r.Uint32()) },
		func() { t.Log(r.Uint64()) },
		func() { t.Log(r.Float32()) },
		func() { t.Log(r.Float64()) },
		func() { t.Log(r.NormFloat64()) },
		func() { t.Log(r.ExpFloat64()) },
		func() {
			n := r.Perm(16)
			for i := 0; i < len(n); i++ {
				require.Less(t, n[i], 16)
			}
		},
		func() {
			r.Shuffle(16, func(i, j int) {
				t.Log(i, j)
				require.Less(t, i, 16)
				require.Less(t, j, 16)
			})
		},
	} {
		wg.Add(1)
		go func(fn func()) {
			defer wg.Done()
			fn()
		}(fn)
	}
	wg.Wait()
}

func BenchmarkNewRand(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		NewRand()
	}

	b.StopTimer()
}

func BenchmarkRand_Int63(b *testing.B) {
	r := NewRand()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r.Int63()
	}

	b.StopTimer()
}

func BenchmarkRand_Bytes(b *testing.B) {
	r := NewRand()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r.Bytes(16)
	}

	b.StopTimer()
}
