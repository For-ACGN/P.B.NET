package lsb

import (
	"bytes"
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"hash"
	"image"
	"image/png"
	"io"
	"math"

	"github.com/pkg/errors"

	"project/internal/convert"
	"project/internal/crypto/aes"
	"project/internal/crypto/hmac"
	"project/internal/security"
)

type pngCommon struct {
	origin   image.Image
	capacity int64
	mode     Mode

	// output/input png image
	nrgba32 *image.NRGBA
	nrgba64 *image.NRGBA64

	// record writer/reader pointer
	current int64
	x       *int
	y       *int
}

func newPNGCommon(img image.Image) *pngCommon {
	rect := img.Bounds()
	width := rect.Dx()
	height := rect.Dy()
	return &pngCommon{
		origin:   img,
		capacity: int64(width * height),
		x:        new(int),
		y:        new(int),
	}
}

// SetOffset is used to set pointer about position.
func (pc *pngCommon) SetOffset(v int64) error {
	if v > pc.Cap() || v < 0 {
		return ErrInvalidOffset
	}
	pc.current = v
	vv := int(v)
	height := pc.origin.Bounds().Dy()
	*pc.x = vv / height
	*pc.y = vv % height
	return nil
}

// Reset is used to reset write or read pointer.
func (pc *pngCommon) Reset() {
	pc.current = 0
	*pc.x = 0
	*pc.y = 0
}

// Image is used to get the original image.
func (pc *pngCommon) Image() image.Image {
	return pc.origin
}

// Cap is used to calculate the capacity that can write or read.
func (pc *pngCommon) Cap() int64 {
	return pc.capacity
}

// Mode is used to get the png writer or reader mode.
func (pc *pngCommon) Mode() Mode {
	return pc.mode
}

// PNGWriter implemented lsb Writer interface.
type PNGWriter struct {
	*pngCommon
}

// NewPNGWriter is used to create a png lsb writer.
func NewPNGWriter(img image.Image, mode Mode) (Writer, error) {
	pw := PNGWriter{
		newPNGCommon(img),
	}
	switch mode {
	case PNGWithNRGBA32:
		pw.nrgba32 = copyNRGBA32(img)
	case PNGWithNRGBA64:
		pw.nrgba64 = copyNRGBA64(img)
	default:
		return nil, errors.New("png writer with " + mode.String())
	}
	pw.mode = mode
	return &pw, nil
}

// Write is used to write data to this image, it will change the under image.
func (pw *PNGWriter) Write(b []byte) (int, error) {
	l := len(b)
	if l == 0 {
		return 0, nil
	}
	ll := int64(l)
	if ll > pw.capacity-pw.current {
		return 0, ErrNoEnoughCapacity
	}
	switch pw.mode {
	case PNGWithNRGBA32:
		writeNRGBA32(pw.origin, pw.nrgba32, pw.x, pw.y, b)
	case PNGWithNRGBA64:
		writeNRGBA64(pw.origin, pw.nrgba64, pw.x, pw.y, b)
	default:
		panic("lsb: internal error")
	}
	pw.current += ll
	return l, nil
}

// Encode is used to encode png to writer.
func (pw *PNGWriter) Encode(w io.Writer) error {
	switch pw.mode {
	case PNGWithNRGBA32:
		return png.Encode(w, pw.nrgba32)
	case PNGWithNRGBA64:
		return png.Encode(w, pw.nrgba64)
	default:
		panic("lsb: internal error")
	}
}

// Reset is used to reset writer and under image.
func (pw *PNGWriter) Reset() {
	pw.pngCommon.Reset()
	switch pw.mode {
	case PNGWithNRGBA32:
		pw.nrgba32 = copyNRGBA32(pw.origin)
	case PNGWithNRGBA64:
		pw.nrgba64 = copyNRGBA64(pw.origin)
	default:
		panic("lsb: internal error")
	}
}

// PNGReader implemented lsb Reader interface.
type PNGReader struct {
	*pngCommon
}

// NewPNGReader is used to create a png lsb reader.
func NewPNGReader(img []byte) (Reader, error) {
	p, err := png.Decode(bytes.NewReader(img))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	pr := PNGReader{
		newPNGCommon(p),
	}
	switch pic := p.(type) {
	case *image.NRGBA:
		pr.mode = PNGWithNRGBA32
		pr.nrgba32 = pic
	case *image.NRGBA64:
		pr.mode = PNGWithNRGBA64
		pr.nrgba64 = pic
	default:
		return nil, errors.Errorf("unsupported png format: %T", pic)
	}
	return &pr, nil
}

// Read is used to read data from png.
func (pr *PNGReader) Read(b []byte) (int, error) {
	l := len(b)
	if l == 0 {
		return 0, nil
	}
	// calculate remaining
	r := pr.capacity - pr.current
	if r <= 0 {
		return 0, io.EOF
	}
	ll := int64(l)
	if r < ll {
		b = b[:r]
		l = int(r)
		ll = r
	}
	switch pr.mode {
	case PNGWithNRGBA32:
		readNRGBA32(pr.nrgba32, pr.x, pr.y, b)
	case PNGWithNRGBA64:
		readNRGBA64(pr.nrgba64, pr.x, pr.y, b)
	default:
		panic("lsb: internal error")
	}
	pr.current += ll
	return l, nil
}

// data structure in png image
//
// +-------------+-------------+----------+-------------+
// | size(int64) | HMAC-SHA256 |    IV    | cipher data |
// +-------------+-------------+----------+-------------+
// |   8 bytes   |  32 bytes   | 16 bytes |     var     |
// +-------------+-------------+----------+-------------+

const (
	pngDataLenSize = convert.Int64Size
	pngReverseSize = pngDataLenSize + sha256.Size + aes.IVSize
)

// PNGEncrypter is used to encrypt data and write it to a png image.
type PNGEncrypter struct {
	writer   Writer
	capacity int64

	hmac hash.Hash
	ctr  *aes.CTR
	iv   *security.Bytes

	offset  int64
	written int64
}

// NewPNGEncrypter is used to create a new png encrypter.
func NewPNGEncrypter(img image.Image, mode Mode, key []byte) (Encrypter, error) {
	writer, err := NewPNGWriter(img, mode)
	if err != nil {
		return nil, err
	}
	// calculate capacity that can encrypt
	capacity := writer.Cap() - pngReverseSize
	if capacity < 1 {
		return nil, ErrImgTooSmall
	}
	pe := PNGEncrypter{
		writer:   writer,
		capacity: capacity,
		hmac:     hmac.New(sha256.New, key),
	}
	err = pe.reset(key, 0)
	if err != nil {
		return nil, err
	}
	return &pe, nil
}

// Write is used to encrypt data and save it to the under image.
func (pe *PNGEncrypter) Write(b []byte) (int, error) {
	l := len(b)
	if l == 0 {
		return 0, nil
	}
	ll := int64(len(b))
	if ll > pe.capacity-pe.offset-pe.written {
		return 0, ErrNoEnoughCapacity
	}
	// encrypt
	cipherData := make([]byte, l)
	pe.ctr.XORKeyStream(cipherData, b)
	// write to image
	n, err := pe.writer.Write(cipherData)
	if err != nil {
		return 0, err
	}
	pe.hmac.Write(cipherData)
	pe.written += ll
	return n, nil
}

// Encode is used to encode under image to writer.
func (pe *PNGEncrypter) Encode(w io.Writer) error {
	if pe.written > 0 {
		err := pe.writeHeader()
		if err != nil {
			return err
		}
	}
	return pe.writer.Encode(w)
}

func (pe *PNGEncrypter) writeHeader() error {
	size := convert.BEInt64ToBytes(pe.written)
	iv := pe.iv.Get()
	defer pe.iv.Put(iv)
	// calculate signature
	pe.hmac.Write(iv)
	pe.hmac.Write(size)
	signature := pe.hmac.Sum(nil)
	// set offset for write header
	err := pe.writer.SetOffset(pe.offset)
	if err != nil {
		panic("lsb: internal error")
	}
	// write header data
	for _, b := range [][]byte{
		size, signature, iv,
	} {
		_, err = pe.writer.Write(b)
		if err != nil {
			return err
		}
	}
	return nil
}

// SetOffset is used to set data start area.
func (pe *PNGEncrypter) SetOffset(v int64) error {
	if pe.written > 0 {
		err := pe.writeHeader()
		if err != nil {
			return err
		}
	}
	return pe.setOffset(v)
}

func (pe *PNGEncrypter) setOffset(offset int64) error {
	err := pe.writer.SetOffset(offset + pngReverseSize)
	if err != nil {
		return err
	}
	iv, err := aes.GenerateIV()
	if err != nil {
		return err
	}
	err = pe.ctr.SetStream(iv)
	if err != nil {
		return err
	}
	pe.hmac.Reset()
	pe.iv = security.NewBytes(iv)
	pe.offset = offset
	pe.written = 0
	return nil
}

// Reset is used to reset png encrypter.
func (pe *PNGEncrypter) Reset(key []byte) error {
	pe.writer.Reset()
	return pe.reset(key, 0)
}

func (pe *PNGEncrypter) reset(key []byte, offset int64) error {
	if key != nil {
		ctr, err := aes.NewCTR(key)
		if err != nil {
			return err
		}
		pe.ctr = ctr
		pe.hmac = hmac.New(sha256.New, key)
	}
	return pe.setOffset(offset)
}

// Key is used to get the aes key.
func (pe *PNGEncrypter) Key() []byte {
	return pe.ctr.Key()
}

// Image is used to get the original png image.
func (pe *PNGEncrypter) Image() image.Image {
	return pe.writer.Image()
}

// Cap is used to calculate the capacity that can encrypt to this png image.
func (pe *PNGEncrypter) Cap() int64 {
	return pe.capacity
}

// Mode is used to get the encrypter mode.
func (pe *PNGEncrypter) Mode() Mode {
	return pe.writer.Mode()
}

// PNGDecrypter is used to read data from a png image and decrypt it.
type PNGDecrypter struct {
	reader   Reader
	capacity int64

	hmac hash.Hash
	ctr  *aes.CTR

	offset int64
	size   int64
	read   int64
}

// NewPNGDecrypter is used to create a new png decrypter.
func NewPNGDecrypter(img, key []byte) (Decrypter, error) {
	reader, err := NewPNGReader(img)
	if err != nil {
		return nil, err
	}
	// calculate capacity that can decrypt
	capacity := reader.Cap() - pngReverseSize
	if capacity < 1 {
		return nil, ErrImgTooSmall
	}
	pd := PNGDecrypter{
		reader:   reader,
		capacity: capacity,
		hmac:     hmac.New(sha256.New, key),
	}
	err = pd.Reset(key)
	if err != nil {
		return nil, err
	}
	return &pd, nil
}

// Read is used to read data and decrypt it.
func (pd *PNGDecrypter) Read(b []byte) (int, error) {
	// if b is nil, can validate only
	if pd.size < 1 {
		err := pd.validate()
		if err != nil {
			return 0, err
		}
	}
	l := int64(len(b))
	if l == 0 {
		return 0, nil
	}
	// calculate remaining
	r := pd.size - pd.read
	if r <= 0 {
		return 0, io.EOF
	}
	if r < l {
		b = b[:r]
		l = r
	}
	// read cipher data
	n, err := io.ReadFull(pd.reader, b)
	if err != nil {
		return 0, err
	}
	// decrypt
	pd.ctr.XORKeyStream(b, b)
	pd.read += l
	return n, nil
}

func (pd *PNGDecrypter) validate() error {
	// read cipher data size
	sizeBuf := make([]byte, pngDataLenSize)
	_, err := io.ReadFull(pd.reader, sizeBuf)
	if err != nil {
		return errors.WithMessage(err, "failed to read cipher data size")
	}
	size := convert.BEBytesToInt64(sizeBuf)
	if size < 1 {
		return errors.New("invalid cipher data size")
	}
	// read HMAC signature
	signature := make([]byte, sha256.Size)
	_, err = io.ReadFull(pd.reader, signature)
	if err != nil {
		return errors.WithMessage(err, "failed to read hmac signature")
	}
	// read iv
	iv := make([]byte, aes.IVSize)
	_, err = io.ReadFull(pd.reader, iv)
	if err != nil {
		return errors.WithMessage(err, "failed to read hmac signature")
	}
	// compare signature
	_, err = io.CopyN(pd.hmac, pd.reader, size)
	if err != nil {
		return errors.WithMessage(err, "failed to read cipher data")
	}
	pd.hmac.Write(iv)
	pd.hmac.Write(sizeBuf)
	if !hmac.Equal(signature, pd.hmac.Sum(nil)) {
		return errors.New("invalid hmac signature")
	}
	// recover offset
	err = pd.reader.SetOffset(pd.offset + pngReverseSize)
	if err != nil {
		return errors.WithMessage(err, "failed to reset offset")
	}
	// set stream
	err = pd.ctr.SetStream(iv)
	if err != nil {
		return err
	}
	pd.size = size
	return nil
}

// SetOffset is used to set data start area.
func (pd *PNGDecrypter) SetOffset(v int64) error {
	return pd.reset(v)
}

// Reset is used to reset png decrypter.
func (pd *PNGDecrypter) Reset(key []byte) error {
	if key != nil {
		ctr, err := aes.NewCTR(key)
		if err != nil {
			return err
		}
		pd.ctr = ctr
		pd.hmac = hmac.New(sha256.New, key)
	}
	return pd.reset(0)
}

func (pd *PNGDecrypter) reset(offset int64) error {
	err := pd.reader.SetOffset(offset)
	if err != nil {
		return err
	}
	pd.hmac.Reset()
	pd.offset = offset
	pd.size = 0
	pd.read = 0
	return nil
}

// Key is used to get the aes key.
func (pd *PNGDecrypter) Key() []byte {
	return pd.ctr.Key()
}

// Image is used to get the original png image.
func (pd *PNGDecrypter) Image() image.Image {
	return pd.reader.Image()
}

// Cap is used to calculate the capacity that can decrypt from this png image.
func (pd *PNGDecrypter) Cap() int64 {
	return pd.capacity
}

// Mode is used to get the decrypter mode.
func (pd *PNGDecrypter) Mode() Mode {
	return pd.reader.Mode()
}

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

	newImg := image.NewNRGBA64(rect)
	x := 0
	y := 0

	writeNRGBA64(img, newImg, &x, &y, secret)

	return newImg, nil
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
	header := make([]byte, headerSize)
	readNRGBA64(img, x, y, header)
	cipherDataSize := int(convert.BEBytesToUint32(header))
	if headerSize+sha256.Size+cipherDataSize > maxSize {
		return nil, errors.New("invalid size in header")
	}
	// read hash
	rawHash := make([]byte, sha256.Size)
	readNRGBA64(img, x, y, rawHash)
	// read cipher data

	cipherData := make([]byte, cipherDataSize)
	readNRGBA64(img, x, y, cipherData)
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
