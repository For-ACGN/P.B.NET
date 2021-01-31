package random

import (
	"crypto/sha256"
	"fmt"
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
	str := String(4096)
	require.Len(t, str, 4096)

	str = String(-1)
	require.Len(t, str, 0)
}

func TestRand_Bool(t *testing.T) {
	fmt.Println(Bool())
}

func TestRand_Intn(t *testing.T) {
	i := Intn(10)
	t.Log(i)
	require.True(t, i >= 0 && i < 10)

	require.True(t, Intn(-1) == 0)
}

func TestRand_Int31n(t *testing.T) {
	i := Int31n(10)
	t.Log(i)
	require.True(t, i >= 0 && i < 10)

	require.True(t, Int31n(-1) == 0)
}

func TestRand_Int63n(t *testing.T) {
	i := Int63n(10)
	t.Log(i)
	require.True(t, i >= 0 && i < 10)

	require.True(t, Int63n(-1) == 0)
}

func TestRand_Number(t *testing.T) {
	t.Log(Int())
	t.Log(Int31())
	t.Log(Int63())

	t.Log(Uint32())
	t.Log(Uint64())

	t.Log(Float32())
	t.Log(Float64())

	t.Log(NormFloat64())
	t.Log(ExpFloat64())
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
		func() { t.Log(r.Int31n(16)) },
		func() { t.Log(r.Int63n(16)) },
		func() { t.Log(r.Int()) },
		func() { t.Log(r.Int31()) },
		func() { t.Log(r.Int63()) },
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
}

func BenchmarkRand_Int63(b *testing.B) {
	r := NewRand()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r.Int63()
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
