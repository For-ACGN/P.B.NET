package lsb

import (
	"image"
	"image/color"

	"project/internal/random"
)

// NRGBA64 is uint16 for each pixel, split one byte to 8 bits
// and save it to it to the the last bit in each color channel(byte).
//
// R: 1111 000[bit1] 1111 000[bit2]
// G: 1111 000[bit3] 1111 000[bit4]
// B: 1111 000[bit5] 1111 000[bit6]
// A: 1111 000[bit6] 1111 000[bit7]

func copyNRGBA64(src image.Image) *image.NRGBA64 {
	rect := src.Bounds()
	min := rect.Min
	width := rect.Dx()
	height := rect.Dy()
	rand := random.NewRand()
	dst := image.NewNRGBA64(rect)
	var rgba color.NRGBA64
	for x := min.X; x < width; x++ {
		for y := min.Y; y < height; y++ {
			rgba = color.NRGBA64Model.Convert(src.At(x, y)).(color.NRGBA64)
			// confuse alpha channel
			switch {
			case rgba.A <= 256:
				if rand.Bool() {
					rgba.A += uint16(rand.Intn(256))
				}
			case rgba.A >= 65535-256:
				if rand.Bool() {
					rgba.A -= uint16(rand.Intn(256))
				}
			default:
				if rand.Bool() {
					rgba.A += uint16(rand.Intn(256))
				} else {
					rgba.A -= uint16(rand.Intn(256))
				}
			}
			dst.SetNRGBA64(x, y, rgba)
		}
	}
	return dst
}

func writeNRGBA64(origin image.Image, img *image.NRGBA64, x, y *int, data []byte) {
	rect := origin.Bounds()
	width := rect.Dx()
	height := rect.Dy()

	var (
		rgba  color.NRGBA64
		block [8]uint8
		byt   byte
		bit   byte
	)

	for i := 0; i < len(data); i++ {
		if *x >= width {
			panic("lsb: out of bounds")
		}

		// write 8 bit to the last bit about 4(RGBA) * 2(front and end) byte
		rgba = color.NRGBA64Model.Convert(origin.At(*x, *y)).(color.NRGBA64)
		block[0] = uint8(rgba.R >> 8) // front 8 bit
		block[1] = uint8(rgba.R)      // end 8 bit
		block[2] = uint8(rgba.G >> 8) // front 8 bit
		block[3] = uint8(rgba.G)      // end 8 bit
		block[4] = uint8(rgba.B >> 8) // front 8 bit
		block[5] = uint8(rgba.B)      // end 8 bit
		block[6] = uint8(rgba.A >> 8) // front 8 bit
		block[7] = uint8(rgba.A)      // end 8 bit

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
		rgba.R = uint16(block[0])<<8 + uint16(block[1])
		rgba.G = uint16(block[2])<<8 + uint16(block[3])
		rgba.B = uint16(block[4])<<8 + uint16(block[5])
		rgba.A = uint16(block[6])<<8 + uint16(block[7])
		img.SetNRGBA64(*x, *y, rgba)

		// check if need go to the next pixel column.
		*y++
		if *y >= height {
			*y = 0
			*x++
		}
	}
}

func readNRGBA64(img *image.NRGBA64, x, y *int, b []byte) {
	rect := img.Bounds()
	width := rect.Dx()
	height := rect.Dy()

	var (
		rgba  color.NRGBA64
		block [8]uint8
		byt   byte
	)

	for i := 0; i < len(b); i++ {
		if *x >= width {
			panic("lsb: out of bounds")
		}

		// read 8 bit to from last bit about 4(RGBA) * 2(front and end) byte
		rgba = img.NRGBA64At(*x, *y)
		block[0] = uint8(rgba.R >> 8) // front 8 bit
		block[1] = uint8(rgba.R)      // end 8 bit
		block[2] = uint8(rgba.G >> 8) // front 8 bit
		block[3] = uint8(rgba.G)      // end 8 bit
		block[4] = uint8(rgba.B >> 8) // front 8 bit
		block[5] = uint8(rgba.B)      // end 8 bit
		block[6] = uint8(rgba.A >> 8) // front 8 bit
		block[7] = uint8(rgba.A)      // end 8 bit

		// set byte
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
	}
}
