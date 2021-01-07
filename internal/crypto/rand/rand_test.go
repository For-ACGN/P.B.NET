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
		n, err := Read(b1)
		require.NoError(t, err)
		require.Equal(t, size, n)

		b2 := make([]byte, size)
		n, err = Read(b2)
		require.NoError(t, err)
		require.Equal(t, size, n)

		require.NotEqual(t, b1, b2)
	})

	t.Run("failed", func(t *testing.T) {
		patch := func(io.Reader, []byte) (int, error) {
			return 0, monkey.Error
		}
		pg := monkey.Patch(io.ReadFull, patch)
		defer pg.Unpatch()

		_, err := Reader.Read(make([]byte, 1024))
		monkey.IsMonkeyError(t, err)
	})
}
