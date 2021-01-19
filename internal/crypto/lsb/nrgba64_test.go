package lsb

import (
	"image"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/random"
	"project/internal/testsuite"
)

func TestWriteNRGBA64(t *testing.T) {
	t.Run("common", func(t *testing.T) {
		testdata := random.Bytes(188)

		origin := image.NewNRGBA64(image.Rect(0, 0, 160, 90))
		img := image.NewNRGBA64(image.Rect(0, 0, 160, 90))
		x := 0
		y := 0

		writeNRGBA64(origin, img, &x, &y, testdata)

		require.Equal(t, 2, x)
		require.Equal(t, 8, y)
	})

	t.Run("panic", func(t *testing.T) {
		origin := image.NewNRGBA64(image.Rect(0, 0, 160, 90))
		img := image.NewNRGBA64(image.Rect(0, 0, 160, 90))
		x := 161
		y := 0

		defer testsuite.DeferForPanic(t)
		writeNRGBA64(origin, img, &x, &y, make([]byte, 4))
	})
}

func TestReadNRGBA64(t *testing.T) {
	t.Run("common", func(t *testing.T) {
		testdata := random.Bytes(188)

		origin := image.NewNRGBA64(image.Rect(0, 0, 160, 90))
		img := image.NewNRGBA64(image.Rect(0, 0, 160, 90))
		x := 0
		y := 0

		writeNRGBA64(origin, img, &x, &y, testdata)

		x = 0
		y = 0
		buf := make([]byte, 188)

		readNRGBA64(img, &x, &y, buf)

		require.Equal(t, testdata, buf)
	})

	t.Run("panic", func(t *testing.T) {
		img := image.NewNRGBA64(image.Rect(0, 0, 160, 90))
		x := 161
		y := 0

		defer testsuite.DeferForPanic(t)
		readNRGBA64(img, &x, &y, make([]byte, 4))
	})
}
