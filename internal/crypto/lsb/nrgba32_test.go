package lsb

import (
	"image"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/random"
	"project/internal/testsuite"
)

func TestWriteNRGBA32(t *testing.T) {
	t.Run("common", func(t *testing.T) {
		testdata := random.Bytes(188)

		origin := image.NewNRGBA(image.Rect(0, 0, 160, 90))
		img := image.NewNRGBA(image.Rect(0, 0, 160, 90))
		x := 0
		y := 0

		writeNRGBA32(origin, img, &x, &y, testdata)

		require.Equal(t, 2, x)
		require.Equal(t, 8, y)
	})

	t.Run("panic", func(t *testing.T) {
		defer testsuite.DeferForPanic(t)

		origin := image.NewNRGBA(image.Rect(0, 0, 160, 90))
		img := image.NewNRGBA(image.Rect(0, 0, 160, 90))
		x := 161
		y := 0

		writeNRGBA32(origin, img, &x, &y, make([]byte, 4))
	})
}

func TestReadNRGBA32(t *testing.T) {
	t.Run("common", func(t *testing.T) {
		testdata := random.Bytes(188)

		origin := image.NewNRGBA(image.Rect(0, 0, 160, 90))
		img := image.NewNRGBA(image.Rect(0, 0, 160, 90))
		x := 0
		y := 0

		writeNRGBA32(origin, img, &x, &y, testdata)

		x = 0
		y = 0
		buf := make([]byte, 188)

		readNRGBA32(img, &x, &y, buf)

		require.Equal(t, testdata, buf)
	})

	t.Run("panic", func(t *testing.T) {
		defer testsuite.DeferForPanic(t)

		img := image.NewNRGBA(image.Rect(0, 0, 160, 90))
		x := 161
		y := 0

		readNRGBA32(img, &x, &y, make([]byte, 4))
	})
}
