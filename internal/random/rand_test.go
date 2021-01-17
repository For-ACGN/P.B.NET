package random

import (
	"crypto/sha256"
	"math/rand"
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

func TestRand(t *testing.T) {
	t.Run("Bytes", func(t *testing.T) {
		bytes := Bytes(10)
		require.Len(t, bytes, 10)
		t.Log(bytes)

		bytes = Bytes(-1)
		require.Len(t, bytes, 0)
	})

	t.Run("String", func(t *testing.T) {
		str := String(10)
		require.Len(t, str, 10)
		t.Log(str)

		str = String(-1)
		require.Len(t, str, 0)
	})

	t.Run("Int", func(t *testing.T) {
		i := Intn(10)
		require.True(t, i >= 0 && i < 10)
		t.Log(i)

		t.Log(Int63())
		t.Log(Uint64())

		require.True(t, Intn(-1) == 0)
	})

	t.Run("panic about rand.New 1", func(t *testing.T) {
		patch := func(rand.Source) *rand.Rand {
			panic(monkey.Panic)
		}
		pg := monkey.Patch(rand.New, patch)
		defer pg.Unpatch()

		defer testDeferForPanic(t)
		NewRand()
	})

	t.Run("panic about rand.New 2", func(t *testing.T) {
		defer time.Sleep(2 * time.Second)

		hash := sha256.New()
		patch := func(interface{}, []byte) (int, error) {
			panic(monkey.Panic)
		}
		pg := monkey.PatchInstanceMethod(hash, "Write", patch)
		defer pg.Unpatch()

		defer testDeferForPanic(t)
		NewRand()
	})
}

func TestRandEqual(t *testing.T) {
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

// performance: BenchmarkNew-6    4148    304633 ns/op    35511 B/op

func BenchmarkNewRand(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		NewRand()
	}
}

func BenchmarkRand_Bytes(b *testing.B) {
	r := NewRand()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r.Bytes(16)
	}
}
