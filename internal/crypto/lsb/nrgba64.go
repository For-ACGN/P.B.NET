package lsb

import (
	"image"
	"image/color"
	"math"
)

// NRGBA64 is uint16 for each pixel, split one byte to 8 bits, and save it to uint16.
//
// R: 1111 000[bit1] 1111 000[bit2]
// G: 1111 000[bit3] 1111 000[bit4]
// B: 1111 000[bit5] 1111 000[bit6]
// A: 1111 000[bit6] 1111 000[bit7]

func encodeNRGBA64(origin image.Image, data []byte) *image.NRGBA64 {
	rect := origin.Bounds()
	min := rect.Min
	max := rect.Max
	width := rect.Dx()
	height := rect.Dy()
	begin := 0
	end := len(data)
	rgba := color.NRGBA64{}
	block := [8]uint8{}
	newImg := image.NewNRGBA64(rect)

	for x := min.X; x < width; x++ {
		for y := min.Y; y < height; y++ {
			r, g, b, a := origin.At(x, y).RGBA()
			rgba.R = uint16(r)
			rgba.G = uint16(g)
			rgba.B = uint16(b)
			rgba.A = uint16(a)

			if begin < end {
				b := data[begin]

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
				for i := 0; i < 8; i++ {
					// get the bit about the byte
					bit := b << i >> 7 // b << (i + 1 - 1) >> 7
					switch {
					case bit == 0 && block[i]&1 == 1:
						block[i]--
					case bit == 1 && block[i]&1 == 0:
						block[i]++
					}
				}

				rgba.R = uint16(block[0])<<8 + uint16(block[1])
				rgba.G = uint16(block[2])<<8 + uint16(block[3])
				rgba.B = uint16(block[4])<<8 + uint16(block[5])
				rgba.A = uint16(block[6])<<8 + uint16(block[7])

				begin++
			} else { // confuse remaining pixel
				switch rgba.A {
				case math.MaxUint16, 0:
				default:
					rgba.A++
				}
			}

			newImg.SetNRGBA64(x, y, rgba)
		}
	}

	// force set the last pixel to make sure image is 64 bit png.
	r, g, b, _ := origin.At(max.X-1, max.Y-1).RGBA()
	c := color.NRGBA64{
		R: uint16(r),
		G: uint16(g),
		B: uint16(b),
		A: 65534,
	}
	newImg.SetNRGBA64(max.X-1, max.Y-1, c)
	return newImg
}

func decodeNRGBA64(img *image.NRGBA64, size int, x, y *int) []byte {
	data := make([]byte, size)
	rect := img.Bounds()
	width := rect.Dx()
	height := rect.Dy()

	var (
		rgba  color.NRGBA64
		block [8]uint8
		b     byte
	)

	for i := 0; i < size; i++ {
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
		for i := 0; i < 8; i++ {
			b += block[i] & 1 << (7 - i)
		}
		data[i] = b

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
	return data
}
