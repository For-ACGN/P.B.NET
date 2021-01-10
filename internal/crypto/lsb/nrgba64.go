package lsb

import (
	"image"
	"image/color"
)

// NRGBA64 is uint16 for each pixel, split one byte to 8 bits
// and save it to it to the the last bit in each color channel(byte).
//
// R: 1111 000[bit1] 1111 000[bit2]
// G: 1111 000[bit3] 1111 000[bit4]
// B: 1111 000[bit5] 1111 000[bit6]
// A: 1111 000[bit6] 1111 000[bit7]

func copyNRGBA64(img image.Image) *image.NRGBA64 {
	rect := img.Bounds()
	min := rect.Min
	width := rect.Dx()
	height := rect.Dy()
	newImg := image.NewNRGBA64(rect)
	var (
		r, g, b, a uint32
		rgba       color.NRGBA64
	)
	for x := min.X; x < width; x++ {
		for y := min.Y; y < height; y++ {
			r, g, b, a = img.At(x, y).RGBA()
			rgba.R = uint16(r)
			rgba.G = uint16(g)
			rgba.B = uint16(b)
			rgba.A = uint16(a)
			newImg.SetNRGBA64(x, y, rgba)
		}
	}
	return newImg
}

func writeNRGBA64(origin image.Image, img *image.NRGBA64, data []byte, x, y *int) {
	rect := origin.Bounds()
	width := rect.Dx()
	height := rect.Dy()

	var (
		r, g, b, a uint32
		rgba       color.NRGBA64
		block      [8]uint8
		byt        byte
		bit        byte
	)

	for i := 0; i < len(data); i++ {
		r, g, b, a = origin.At(*x, *y).RGBA()
		rgba.R = uint16(r)
		rgba.G = uint16(g)
		rgba.B = uint16(b)
		rgba.A = uint16(a)

		// write 8 bit to the last bit about 4(RGBA) * 2(front and end) byte
		block[0] = uint8(rgba.R >> 8) // red front 8 bit
		block[1] = uint8(rgba.R)      // red end 8 bit
		block[2] = uint8(rgba.G >> 8) // green front 8 bit
		block[3] = uint8(rgba.G)      // green end 8 bit
		block[4] = uint8(rgba.B >> 8) // blue front 8 bit
		block[5] = uint8(rgba.B)      // blue end 8 bit
		block[6] = uint8(rgba.A >> 8) // alpha front 8 bit
		block[7] = uint8(rgba.A)      // alpha end 8 bit

		// update original pixel
		byt = data[i]
		for j := 0; j < 8; j++ {
			// get the bit about the byte
			bit = byt << j >> 7 // b << (j + 1 - 1) >> 7
			switch {
			case bit == 0 && block[j]&1 == 1:
				block[j]--
			case bit == 1 && block[j]&1 == 0:
				block[j]++
			}
			// reset bit
			bit = 0
		}

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
		if *x >= width {
			panic("lsb: internal error")
		}
	}
}

func readNRGBA64(img *image.NRGBA64, b []byte, x, y *int) {
	rect := img.Bounds()
	width := rect.Dx()
	height := rect.Dy()

	var (
		rgba  color.NRGBA64
		block [8]uint8
		byt   byte
	)

	for i := 0; i < len(b); i++ {
		rgba = img.NRGBA64At(*x, *y)

		// write 8 bit to the last bit about 4(RGBA) * 2(front and end) byte
		block[0] = uint8(rgba.R >> 8) // red front 8 bit
		block[1] = uint8(rgba.R)      // red end 8 bit
		block[2] = uint8(rgba.G >> 8) // green front 8 bit
		block[3] = uint8(rgba.G)      // green end 8 bit
		block[4] = uint8(rgba.B >> 8) // blue front 8 bit
		block[5] = uint8(rgba.B)      // blue end 8 bit
		block[6] = uint8(rgba.A >> 8) // alpha front 8 bit
		block[7] = uint8(rgba.A)      // alpha end 8 bit

		// set byte
		for j := 0; j < 8; j++ {
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
			panic("lsb: internal error")
		}
	}
}
