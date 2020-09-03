package security

import (
	"reflect"
	"strings"
	"sync"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"
)

func TestCoverBytes(t *testing.T) {
	b1 := []byte{1, 2, 3, 4}
	b2 := []byte{1, 2, 3, 4}
	CoverBytes(b2)
	require.NotEqual(t, b1, b2, "failed to cover bytes")
}

func TestCoverString(t *testing.T) {
	// must use strings.Repeat to generate testdata
	// if you use this
	// s1 := "aaa"
	// s2 := "aaa"
	// CoverString(&s1) will panic, because it change const.

	s1 := strings.Repeat("a", 10)
	s2 := strings.Repeat("a", 10)
	CoverString(s2)
	require.NotEqual(t, s1, s2, "failed to cover string")
}

func TestBytes(t *testing.T) {
	testdata := []byte{1, 2, 3, 4}

	sb := NewBytes(testdata)

	t.Run("common", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			b := sb.Get()
			require.Equal(t, testdata, b)
			sb.Put(b)
		}
	})

	t.Run("compare address", func(t *testing.T) {
		// maybe gc, so try 1000 times
		var equal bool
		for i := 0; i < 1000; i++ {
			b := sb.Get()
			addr1 := (*reflect.SliceHeader)(unsafe.Pointer(&b)).Data
			sb.Put(b)

			b = sb.Get()
			addr2 := (*reflect.SliceHeader)(unsafe.Pointer(&b)).Data
			sb.Put(b)

			if addr1 == addr2 {
				equal = true
				break
			}
		}
		require.True(t, equal)
	})

	t.Run("parallel", func(t *testing.T) {
		wg := sync.WaitGroup{}
		wg.Add(100)
		for i := 0; i < 100; i++ {
			go func() {
				defer wg.Done()
				for i := 0; i < 10; i++ {
					b := sb.Get()
					require.Equal(t, testdata, b)
					sb.Put(b)
				}
			}()
		}
		wg.Wait()
	})
}

func BenchmarkBytes(b *testing.B) {
	b.Run("32 bytes", func(b *testing.B) {
		benchmarkBytes(b, 32)
	})

	b.Run("64 bytes", func(b *testing.B) {
		benchmarkBytes(b, 64)
	})

	b.Run("128 bytes", func(b *testing.B) {
		benchmarkBytes(b, 128)
	})
}

func benchmarkBytes(b *testing.B, n int) {
	sb := NewBytes(make([]byte, n))

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		data := sb.Get()
		sb.Put(data)
	}
}
