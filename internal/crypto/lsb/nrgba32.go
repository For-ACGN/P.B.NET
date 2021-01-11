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

func copyNRGBA32(src image.Image) *image.NRGBA {
	rect := src.Bounds()
	min := rect.Min
	width := rect.Dx()
	height := rect.Dy()
	dst := image.NewNRGBA(rect)
	var (
		r, g, b, a uint32
		rgba       color.NRGBA
	)
	for x := min.X; x < width; x++ {
		for y := min.Y; y < height; y++ {
			r, g, b, a = src.At(x, y).RGBA()
			rgba.R = uint8(r)
			rgba.G = uint8(g)
			rgba.B = uint8(b)
			rgba.A = uint8(a)
			dst.SetNRGBA(x, y, rgba)
		}
	}
	return dst
}

func writeNRGBA32(origin image.Image, img *image.NRGBA, data []byte, x, y *int) {
	rect := origin.Bounds()
	width := rect.Dx()
	height := rect.Dy()

	var (
		r, g, b, a uint32
		rgba       color.NRGBA
		block      [8]uint8
		byt        byte
		bit        byte
	)

	for i := 0; i < len(data); i++ {
		r, g, b, a = origin.At(*x, *y).RGBA()
		rgba.R = uint8(r)
		rgba.G = uint8(g)
		rgba.B = uint8(b)
		rgba.A = uint8(a)

		// write 8 bit to the last two and last one bit in each color channel
		block[0] = rgba.R >> 1 // red last two bit
		block[1] = rgba.R      // red last one bit
		block[2] = rgba.G >> 1 // green last two bit
		block[3] = rgba.G      // green last one bit
		block[4] = rgba.B >> 1 // blue last two bit
		block[5] = rgba.B      // blue last one bit
		block[6] = rgba.A >> 1 // alpha last two bit
		block[7] = rgba.A      // alpha last one bit

		// update original pixel
		byt = data[i]
		for j := 0; j < 8; j++ {
			// get each bit about the byte
			bit = byt << j >> 7 // b << (j + 1 - 1) >> 7
			// compare and check need update
			switch {
			case bit == 0 && block[j]&1 == 1:
				block[j]--
			case bit == 1 && block[j]&1 == 0:
				block[j]++
			}
			// reset bit
			bit = 0
		}

		// save the final pixel
		rgba.R = block[0]<<1 + block[1]&1
		rgba.G = block[2]<<1 + block[3]&1
		rgba.B = block[4]<<1 + block[5]&1
		rgba.A = block[6]<<1 + block[7]&1
		img.SetNRGBA(*x, *y, rgba)

		// check if need go to the next pixel column.
		*y++
		if *y >= height {
			*y = 0
			*x++
		}
		if *x >= width {
			panic("lsb: out of bounds")
		}
	}
}

func readNRGBA32(img *image.NRGBA, b []byte, x, y *int) {
	rect := img.Bounds()
	width := rect.Dx()
	height := rect.Dy()

	var (
		rgba  color.NRGBA
		block [8]uint8
		byt   byte
	)

	for i := 0; i < len(b); i++ {
		rgba = img.NRGBAAt(*x, *y)

		// read 8 bit from the last two and last one bit in each color channel
		block[0] = rgba.R >> 1 // red last two bit
		block[1] = rgba.R      // red last one bit
		block[2] = rgba.G >> 1 // green last two bit
		block[3] = rgba.G      // green last one bit
		block[4] = rgba.B >> 1 // blue last two bit
		block[5] = rgba.B      // blue last one bit
		block[6] = rgba.A >> 1 // alpha last two bit
		block[7] = rgba.A      // alpha last one bit

		// set bytes
		for j := 0; j < 8; j++ {
			// get the last bit of this byte
			byt += block[j] & 1 << (7 - j)
		}
		b[i] = byt

		// reset byte
		byt = 0

		// check if need go to the next pixel column.
		*y++
		if *y >= height {
			*y = 0
			*x++
		}
		if *x >= width {
			panic("lsb: out of bounds")
		}
	}
}
