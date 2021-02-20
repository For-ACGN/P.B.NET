package compare

import (
	"fmt"
	"net"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUniqueStrings(t *testing.T) {
	for _, testdata := range [...]*struct {
		previous []string
		current  []string
		added    []int
		deleted  []int
	}{
		{
			previous: []string{"a", "b", "c"},
			current:  []string{"b", "c", "d"},
			added:    []int{2},
			deleted:  []int{0},
		},
	} {
		added, deleted := UniqueStrings(testdata.previous, testdata.current)
		require.Equal(t, testdata.added, added)
		require.Equal(t, testdata.deleted, deleted)
	}
}

func BenchmarkUniqueStrings(b *testing.B) {
	// 2 = port(uint16) size
	// 4 = zone(uint32) size
	const (
		tcp4RowSize = net.IPv4len + 2 + net.IPv4len + 2
		tcp6RowSize = net.IPv6len + 4 + 2 + net.IPv6len + 4 + 2
	)

	b.Run("100 x tcp4RowSize", func(b *testing.B) {
		benchmarkUniqueStrings(b, 100, tcp4RowSize)
	})

	b.Run("1000 x tcp4RowSize", func(b *testing.B) {
		benchmarkUniqueStrings(b, 1000, tcp4RowSize)
	})

	b.Run("10000 x tcp4RowSize", func(b *testing.B) {
		benchmarkUniqueStrings(b, 10000, tcp4RowSize)
	})

	b.Run("100000 x tcp4RowSize", func(b *testing.B) {
		benchmarkUniqueStrings(b, 100000, tcp4RowSize)
	})

	b.Run("100 x tcp6RowSize", func(b *testing.B) {
		benchmarkUniqueStrings(b, 100, tcp6RowSize)
	})

	b.Run("1000 x tcp6RowSize", func(b *testing.B) {
		benchmarkUniqueStrings(b, 1000, tcp6RowSize)
	})

	b.Run("10000 x tcp6RowSize", func(b *testing.B) {
		benchmarkUniqueStrings(b, 10000, tcp6RowSize)
	})

	b.Run("100000 x tcp6RowSize", func(b *testing.B) {
		benchmarkUniqueStrings(b, 100000, tcp6RowSize)
	})
}

func benchmarkUniqueStrings(b *testing.B, size, factor int) {
	n := make([]string, size)
	for i := 0; i < size; i++ {
		n[i] = fmt.Sprintf("%0"+strconv.Itoa(factor)+"d", i+1)
	}
	o := make([]string, size)
	for i := 0; i < size; i++ {
		o[i] = fmt.Sprintf("%0"+strconv.Itoa(factor)+"d", i)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		added, deleted := UniqueStrings(o, n)
		if len(added) != 1 {
			b.Fatal("invalid added number:", added)
		}
		if len(deleted) != 1 {
			b.Fatal("invalid deleted number:", deleted)
		}
		if added[0] != size-1 {
			b.Fatal("invalid added index:", added[0])
		}
		if deleted[0] != 0 {
			b.Fatal("invalid deleted index:", deleted[0])
		}
	}
}
