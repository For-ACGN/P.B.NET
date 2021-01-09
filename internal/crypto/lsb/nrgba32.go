package lsb

import (
	"image"
)

// NRGBA32 is uint8 for each pixel, split one byte to 8 bits
// and save it to the the last two bit in each color channel.
//
// R: 1111 00[bit1][bit2]
// G: 1111 00[bit3][bit4]
// B: 1111 00[bit5][bit6]
// A: 1111 00[bit7][bit8]

func writeNRGBA32(origin image.Image, img *image.NRGBA, data []byte, x, y *int) {

}

func readNRGBA32(img *image.NRGBA, b []byte, x, y *int) []byte {
	return nil
}
