package lsb

import (
	"image"
	"image/color"
)

// NRGBA32 is uint8 for each pixel, split one byte to 8 bits
// and save it to the the last two bit in each color channel.
//
// R: 1111 00[bit1][bit2]
// G: 1111 00[bit3][bit4]
// B: 1111 00[bit5][bit6]
// A: 1111 00[bit7][bit8]

func copyNRGBA32(img image.Image) *image.NRGBA {
	rect := img.Bounds()
	min := rect.Min
	width := rect.Dx()
	height := rect.Dy()
	newImg := image.NewNRGBA(rect)
	var (
		r, g, b, a uint32
		rgba       color.NRGBA
	)
	for x := min.X; x < width; x++ {
		for y := min.Y; y < height; y++ {
			r, g, b, a = img.At(x, y).RGBA()
			rgba.R = uint8(r)
			rgba.G = uint8(g)
			rgba.B = uint8(b)
			rgba.A = uint8(a)
			newImg.SetNRGBA(x, y, rgba)
		}
	}
	return newImg
}

func writeNRGBA32(origin image.Image, img *image.NRGBA, data []byte, x, y *int) {

}

func readNRGBA32(img *image.NRGBA, b []byte, x, y *int) []byte {
	return nil
}
