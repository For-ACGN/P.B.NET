package rand

import (
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/patch/monkey"
)

func TestReader_Read(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		const size = 32

		b1 := make([]byte, size)
		b2 := make([]byte, size)

		for i := 0; i < 1024; i++ {
			n, err := Read(b1)
			require.NoError(t, err)
			require.Equal(t, size, n)

			n, err = Read(b2)
			require.NoError(t, err)
			require.Equal(t, size, n)

			require.NotEqual(t, b1, b2)
		}
	})

	t.Run("failed to read", func(t *testing.T) {
		patch := func(io.Reader, []byte) (int, error) {
			return 0, monkey.Error
		}
		pg := monkey.Patch(io.ReadFull, patch)
		defer pg.Unpatch()

		_, err := Reader.Read(make([]byte, 1024))
		monkey.IsMonkeyError(t, err)
	})
}

func BenchmarkReader_Read(b *testing.B) {
	buf := make([]byte, 16) // AES IV size

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := Read(buf)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.StopTimer()
}
