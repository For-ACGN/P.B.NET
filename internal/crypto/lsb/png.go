package lsb

import (
	"bytes"
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"image"
	"image/png"
	"io"
	"math"

	"github.com/pkg/errors"

	"project/internal/convert"
	"project/internal/crypto/aes"
	"project/internal/security"
)

// PNGWriter implemented lsb Writer interface.
type PNGWriter struct {
	origin image.Image
	mode   Mode

	// output png image
	nrgba32 *image.NRGBA
	nrgba64 *image.NRGBA64

	// record write pointer
	x *int
	y *int
}

// NewPNGWriter is used to create a png lsb writer.
func NewPNGWriter(img image.Image, mode Mode) (*PNGWriter, error) {
	switch mode {
	case PNGWithNRGBA32, PNGWithNRGBA64:
	default:
		return nil, errors.New(mode.String())
	}
	p := PNGWriter{
		origin: img,
		mode:   mode,
		x:      new(int),
		y:      new(int),
	}
	switch mode {
	case PNGWithNRGBA32:
		p.nrgba32 = image.NewNRGBA(img.Bounds())
	case PNGWithNRGBA64:
		p.nrgba64 = image.NewNRGBA64(img.Bounds())
	}
	return &p, nil
}

// StorageSize is used to calculate the size that can write.
func (p *PNGWriter) StorageSize() uint64 {
	rect := p.origin.Bounds()
	width := rect.Dx()
	height := rect.Dy()
	return uint64(width * height)
}

// Image is used to get the inner image that will be encoded.
func (p *PNGWriter) Image() image.Image {
	switch p.mode {
	case PNGWithNRGBA32:
		return &image.NRGBA{
			Pix:    append([]byte{}, p.nrgba32.Pix...),
			Stride: p.nrgba32.Stride,
			Rect:   p.nrgba32.Rect,
		}
	case PNGWithNRGBA64:
		return &image.NRGBA64{
			Pix:    append([]byte{}, p.nrgba64.Pix...),
			Stride: p.nrgba64.Stride,
			Rect:   p.nrgba64.Rect,
		}
	default:
		panic("lsb: internal error")
	}
}

// Write is used to write data to this image, it will change the under image.
func (p *PNGWriter) Write(data []byte) (int, error) {
	switch p.mode {
	case PNGWithNRGBA32:
		writeNRGBA32(p.origin, p.nrgba32, data, p.x, p.y)
		return len(data), nil
	case PNGWithNRGBA64:
		writeNRGBA64(p.origin, p.nrgba64, data, p.x, p.y)
		return len(data), nil
	default:
		panic("lsb: internal error")
	}
}

// Encode is used to encode png to writer.
func (p *PNGWriter) Encode(w io.Writer) error {
	switch p.mode {
	case PNGWithNRGBA32:
		return png.Encode(w, p.nrgba32)
	case PNGWithNRGBA64:
		return png.Encode(w, p.nrgba64)
	default:
		panic("lsb: internal error")
	}
}

// Reset is used to reset writer.
func (p *PNGWriter) Reset() {
	switch p.mode {
	case PNGWithNRGBA32:
		p.nrgba32 = image.NewNRGBA(p.origin.Bounds())
	case PNGWithNRGBA64:
		p.nrgba64 = image.NewNRGBA64(p.origin.Bounds())
	}
	*p.x = 0
	*p.y = 0
}

// Mode is used to get the png writer mode.
func (p *PNGWriter) Mode() Mode {
	return p.mode
}

// data structure stored in PNG
// +--------------+----------+-----------+
// | size(uint32) |  SHA256  | AES(data) |
// +--------------+----------+-----------+
// |   4 bytes    | 32 bytes |    var    |
// +--------------+----------+-----------+

// size is uint32
const headerSize = 4

// CalculateStorageSize is used to calculate the maximum data that can encrypted.
func CalculateStorageSize(rect image.Rectangle) int {
	width := rect.Dx()
	height := rect.Dy()
	size := width * height
	// sha256.Size-1,  "1" is reserved pixel, see encodeNRGBA64()
	block := (size-headerSize-sha256.Size-1)/aes.BlockSize - 1 // "1" is for aes padding
	// actual data that can store
	max := block*aes.BlockSize + (aes.BlockSize - 1)
	if max < 0 {
		max = 0
	}
	return max
}

// EncryptToPNG is used to load PNG image and encrypt data to it.
func EncryptToPNG(pic, plainData, key, iv []byte) ([]byte, error) {
	img, err := png.Decode(bytes.NewReader(pic))
	if err != nil {
		return nil, err
	}
	newImg, err := Encrypt(img, plainData, key, iv)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(make([]byte, 0, len(pic)))
	encoder := png.Encoder{
		CompressionLevel: png.BestCompression,
	}
	err = encoder.Encode(buf, newImg)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Encrypt is used to encrypt data by aes + hash and save it to a PNG image.
func Encrypt(img image.Image, plainData, key, iv []byte) (*image.NRGBA64, error) {
	// basic information
	rect := img.Bounds()
	storageSize := CalculateStorageSize(rect)
	size := len(plainData)
	if size > storageSize {
		const format = "this image can only store %s data, plain data size is %d"
		str := convert.StorageUnit(uint64(storageSize))
		return nil, fmt.Errorf(format, str, size)
	}
	if size > math.MaxInt32-1 { // because aes block size
		return nil, errors.New("plain data size is bigger than 4GB")
	}
	// encrypt data
	cipherData, err := aes.CBCEncrypt(plainData, key)
	if err != nil {
		return nil, err
	}
	defer security.CoverBytes(cipherData)
	h := sha256.Sum256(plainData)
	hash := h[:]
	defer security.CoverBytes(hash)
	// set secret
	secret := make([]byte, 0, headerSize+sha256.Size+len(cipherData))
	secret = append(secret, convert.BEUint32ToBytes(uint32(len(cipherData)))...)
	secret = append(secret, hash...)
	secret = append(secret, cipherData...)
	return encodeNRGBA64_old(img, secret), nil
}

// DecryptFromPNG is used to load a PNG image and  decrypt data from it.
func DecryptFromPNG(pic, key, iv []byte) ([]byte, error) {
	p, err := png.Decode(bytes.NewReader(pic))
	if err != nil {
		return nil, err
	}
	img, ok := p.(*image.NRGBA64)
	if !ok {
		return nil, errors.New("png is not NRGBA64")
	}
	return Decrypt(img, key, iv)
}

// Decrypt is used to decrypt cipher data from a PNG image.
func Decrypt(img *image.NRGBA64, key, iv []byte) ([]byte, error) {
	// basic information
	rect := img.Bounds()
	width, height := rect.Dx(), rect.Dy()
	maxSize := width * height // one pixel one byte
	if maxSize < headerSize+sha256.Size+aes.BlockSize {
		return nil, errors.New("invalid image size")
	}
	min := rect.Min
	// store global position
	x := &min.X
	y := &min.Y
	// read header
	header := readNRGBA64(img, headerSize, x, y)
	cipherDataSize := int(convert.BEBytesToUint32(header))
	if headerSize+sha256.Size+cipherDataSize > maxSize {
		return nil, errors.New("invalid size in header")
	}
	// read hash
	rawHash := readNRGBA64(img, sha256.Size, x, y)
	// read cipher data
	cipherData := readNRGBA64(img, cipherDataSize, x, y)
	// decrypt
	plainData, err := aes.CBCDecrypt(cipherData, key)
	if err != nil {
		return nil, err
	}
	// check hash
	hash := sha256.Sum256(plainData)
	if subtle.ConstantTimeCompare(hash[:], rawHash) != 1 {
		return nil, errors.New("invalid hash about the plain data")
	}
	return plainData, nil
}
