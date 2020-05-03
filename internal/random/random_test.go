package random

import (
	"crypto/sha256"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"project/internal/patch/monkey"
)

func TestRandom(t *testing.T) {
	t.Run("String", func(t *testing.T) {
		s := String(10)
		require.True(t, len(s) == 10)
		t.Log(s)

		require.True(t, len(String(-1)) == 0)
	})

	t.Run("Bytes", func(t *testing.T) {
		b := Bytes(10)
		require.True(t, len(b) == 10)
		t.Log(b)

		require.True(t, len(Bytes(-1)) == 0)
	})

	t.Run("Cookie", func(t *testing.T) {
		c := Cookie(10)
		require.True(t, len(c) == 10)
		t.Log(c)

		require.True(t, len(Cookie(-1)) == 0)
	})

	t.Run("Cookie-collide", func(t *testing.T) {
		for i := 0; i < 10240; i++ {
			Cookie(32)
		}
	})

	t.Run("Int", func(t *testing.T) {
		i := Int(10)
		require.True(t, i >= 0 && i < 10)
		t.Log(i)
		t.Log(Int64())
		t.Log(Uint64())

		require.True(t, Int(-1) == 0)
	})

	t.Run("panic about rand.New 1", func(t *testing.T) {
		defer func() { require.NotNil(t, recover()) }()
		patch := func(_ rand.Source) *rand.Rand {
			panic(monkey.Panic)
		}
		pg := monkey.Patch(rand.New, patch)
		defer pg.Unpatch()
		New()
	})

	t.Run("panic about rand.New 2", func(t *testing.T) {
		defer func() {
			require.NotNil(t, recover())
			time.Sleep(2 * time.Second)
		}()
		hash := sha256.New()
		patch := func(_ interface{}, _ []byte) (int, error) {
			panic(monkey.Panic)
		}
		pg := monkey.PatchInstanceMethod(hash, "Write", patch)
		defer pg.Unpatch()
		New()
	})
}

func TestRandomEqual(t *testing.T) {
	const n = 64
	result := make(chan int, n)
	for i := 0; i < n; i++ {
		go func() {
			r := New()
			result <- r.Int(1048576)
		}()
	}
	results := make(map[int]*struct{})
	for i := 0; i < n; i++ {
		r := <-result
		_, ok := results[r]
		require.False(t, ok, "appeared value: %d, i: %d", r, i)
		results[r] = new(struct{})
	}
}

func BenchmarkNew(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		New()
	}
}

func BenchmarkRand_Bytes(b *testing.B) {
	r := New()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Bytes(16)
	}
}

func TestSleeper(t *testing.T) {
	time.Sleep(100 * time.Millisecond)

	t.Run("common", func(t *testing.T) {
		<-Sleep(1, 2)
	})

	t.Run("zero", func(t *testing.T) {
		<-Sleep(0, 0)
	})

	t.Run("not read", func(t *testing.T) {
		Sleep(0, 0)
		time.Sleep(time.Second + 100*time.Millisecond)
		Sleep(0, 0)
	})

	t.Run("max", func(t *testing.T) {
		d := gSleeper.calculateDuration(3600, 3600)
		require.Equal(t, MaxSleepTime, d)
	})

	gSleeper.Stop()
}
