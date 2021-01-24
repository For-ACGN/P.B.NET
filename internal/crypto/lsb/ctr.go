package lsb

import (
	"crypto/sha256"
	"hash"
	"image"
	"io"

	"github.com/pkg/errors"

	"project/internal/convert"
	"project/internal/crypto/aes"
	"project/internal/crypto/hmac"
	"project/internal/security"
)

// Data structure in the under image.
//
// +-------------+-------------+----------+-------------+
// | size(int64) | HMAC-SHA256 |    IV    | cipher data |
// +-------------+-------------+----------+-------------+
// |   8 bytes   |  32 bytes   | 16 bytes |     var     |
// +-------------+-------------+----------+-------------+

const (
	ctrDataLenSize = convert.Int64Size
	ctrReverseSize = ctrDataLenSize + sha256.Size + aes.IVSize
)

var _ Encrypter = new(CTREncrypter)

// CTREncrypter is used to encrypt data and write it to a lsb Writer.
type CTREncrypter struct {
	writer   Writer
	capacity int64

	hmac hash.Hash
	ctr  *aes.CTR
	iv   *security.Bytes

	offset  int64
	written int64
}

// NewCTREncrypter is used to create a new AES-CTR encrypter.
func NewCTREncrypter(writer Writer, key []byte) (*CTREncrypter, error) {
	// calculate capacity that can encrypt
	capacity := writer.Cap() - ctrReverseSize
	if capacity < 1 {
		return nil, ErrImgTooSmall
	}
	// create encrypter
	pe := CTREncrypter{
		writer:   writer,
		capacity: capacity,
		hmac:     hmac.New(sha256.New, key),
	}
	err := pe.reset(key, 0)
	if err != nil {
		return nil, err
	}
	return &pe, nil
}

// Write is used to encrypt data and save it to the under image.
func (ce *CTREncrypter) Write(b []byte) (int, error) {
	l := len(b)
	if l == 0 {
		return 0, nil
	}
	ll := int64(len(b))
	if ll > ce.capacity-ce.offset-ce.written {
		return 0, ErrNoEnoughCapacity
	}
	// encrypt
	cipherData := make([]byte, l)
	ce.ctr.XORKeyStream(cipherData, b)
	// write to image
	n, err := ce.writer.Write(cipherData)
	if err != nil {
		return 0, err
	}
	ce.hmac.Write(cipherData)
	ce.written += ll
	return n, nil
}

// Encode is used to encode under image to writer.
func (ce *CTREncrypter) Encode(w io.Writer) error {
	if ce.written > 0 {
		err := ce.writeHeader()
		if err != nil {
			return err
		}
	}
	return ce.writer.Encode(w)
}

func (ce *CTREncrypter) writeHeader() error {
	size := convert.BEInt64ToBytes(ce.written)
	// get iv
	iv := ce.iv.Get()
	defer ce.iv.Put(iv)
	// encrypt size buffer
	size, err := ce.ctr.EncryptWithIV(size, iv)
	if err != nil {
		panic("lsb: internal error")
	}
	// calculate signature
	ce.hmac.Write(iv)
	ce.hmac.Write(size)
	signature := ce.hmac.Sum(nil)
	// set offset for write header
	err = ce.writer.SetOffset(ce.offset)
	if err != nil {
		panic("lsb: internal error")
	}
	// write header data
	for _, b := range [][]byte{
		size, signature, iv,
	} {
		_, err = ce.writer.Write(b)
		if err != nil {
			return err
		}
	}
	return nil
}

// SetOffset is used to set data start area.
func (ce *CTREncrypter) SetOffset(v int64) error {
	if v < 0 {
		return ErrInvalidOffset
	}
	if ce.written > 0 {
		err := ce.writeHeader()
		if err != nil {
			return err
		}
	}
	return ce.setOffset(v)
}

func (ce *CTREncrypter) setOffset(offset int64) error {
	err := ce.writer.SetOffset(offset + ctrReverseSize)
	if err != nil {
		return err
	}
	iv, err := aes.GenerateIV()
	if err != nil {
		return err
	}
	err = ce.ctr.SetStream(iv)
	if err != nil {
		return err
	}
	ce.hmac.Reset()
	ce.iv = security.NewBytes(iv)
	ce.offset = offset
	ce.written = 0
	return nil
}

// Reset is used to reset AES-CTR encrypter.
func (ce *CTREncrypter) Reset(key []byte) error {
	ce.writer.Reset()
	return ce.reset(key, 0)
}

func (ce *CTREncrypter) reset(key []byte, offset int64) error {
	if key != nil {
		ctr, err := aes.NewCTR(key)
		if err != nil {
			return err
		}
		ce.ctr = ctr
		ce.hmac = hmac.New(sha256.New, key)
	}
	return ce.setOffset(offset)
}

// Key is used to get the aes key.
func (ce *CTREncrypter) Key() []byte {
	return ce.ctr.Key()
}

// Image is used to get the original image in the under Writer.
func (ce *CTREncrypter) Image() image.Image {
	return ce.writer.Image()
}

// Cap is used to calculate the capacity that can encrypt to the under Writer.
func (ce *CTREncrypter) Cap() int64 {
	return ce.capacity
}

// Mode is used to get the mode about the under Writer.
func (ce *CTREncrypter) Mode() Mode {
	return ce.writer.Mode()
}

// Algorithm is used to get the algorithm.
func (ce *CTREncrypter) Algorithm() Algorithm {
	return AESWithCTR
}

var _ Decrypter = new(CTRDecrypter)

// CTRDecrypter is used to read data from a lsb Reader and decrypt it.
type CTRDecrypter struct {
	reader   Reader
	capacity int64

	hmac hash.Hash
	ctr  *aes.CTR

	offset int64
	size   int64
	read   int64
}

// NewCTRDecrypter is used to create a new AES-CTR decrypter.
func NewCTRDecrypter(reader Reader, key []byte) (*CTRDecrypter, error) {
	// calculate capacity that can decrypt
	capacity := reader.Cap() - ctrReverseSize
	if capacity < 1 {
		return nil, ErrImgTooSmall
	}
	// create decrypter
	pd := CTRDecrypter{
		reader:   reader,
		capacity: capacity,
		hmac:     hmac.New(sha256.New, key),
	}
	err := pd.Reset(key)
	if err != nil {
		return nil, err
	}
	return &pd, nil
}

// Read is used to read data and decrypt it.
func (cd *CTRDecrypter) Read(b []byte) (int, error) {
	// if b is nil, can validate only
	if cd.size < 1 {
		err := cd.validate()
		if err != nil {
			return 0, err
		}
	}
	l := int64(len(b))
	if l == 0 {
		return 0, nil
	}
	// calculate remaining
	r := cd.size - cd.read
	if r <= 0 {
		return 0, io.EOF
	}
	if r < l {
		b = b[:r]
		l = r
	}
	// read cipher data
	n, err := io.ReadFull(cd.reader, b)
	if err != nil {
		return 0, err
	}
	// decrypt
	cd.ctr.XORKeyStream(b, b)
	cd.read += l
	return n, nil
}

func (cd *CTRDecrypter) validate() error {
	// read cipher data size
	sizeBuf := make([]byte, ctrDataLenSize)
	_, err := io.ReadFull(cd.reader, sizeBuf)
	if err != nil {
		return errors.WithMessage(err, "failed to read cipher data size")
	}
	// read HMAC signature
	signature := make([]byte, sha256.Size)
	_, err = io.ReadFull(cd.reader, signature)
	if err != nil {
		return errors.WithMessage(err, "failed to read hmac signature")
	}
	// read iv
	iv := make([]byte, aes.IVSize)
	_, err = io.ReadFull(cd.reader, iv)
	if err != nil {
		return errors.WithMessage(err, "failed to read iv")
	}
	// decrypt size
	sizeBufDec, err := cd.ctr.DecryptWithIV(sizeBuf, iv)
	if err != nil {
		return errors.WithMessage(err, "failed to decrypt buffer about size")
	}
	size := convert.BEBytesToInt64(sizeBufDec)
	if size < 1 {
		return errors.New("invalid cipher data size")
	}
	// compare signature
	_, err = io.CopyN(cd.hmac, cd.reader, size)
	if err != nil {
		return errors.WithMessage(err, "failed to read cipher data")
	}
	cd.hmac.Write(iv)
	cd.hmac.Write(sizeBuf)
	if !hmac.Equal(signature, cd.hmac.Sum(nil)) {
		return errors.New("invalid hmac signature")
	}
	// reset offset
	err = cd.reader.SetOffset(cd.offset + ctrReverseSize)
	if err != nil {
		return errors.WithMessage(err, "failed to reset offset")
	}
	// set stream
	err = cd.ctr.SetStream(iv)
	if err != nil {
		return err
	}
	cd.size = size
	return nil
}

// SetOffset is used to set data start area.
func (cd *CTRDecrypter) SetOffset(v int64) error {
	return cd.reset(v)
}

// Reset is used to reset AES-CTR decrypter.
func (cd *CTRDecrypter) Reset(key []byte) error {
	if key != nil {
		ctr, err := aes.NewCTR(key)
		if err != nil {
			return err
		}
		cd.ctr = ctr
		cd.hmac = hmac.New(sha256.New, key)
	}
	return cd.reset(0)
}

func (cd *CTRDecrypter) reset(offset int64) error {
	err := cd.reader.SetOffset(offset)
	if err != nil {
		return err
	}
	cd.hmac.Reset()
	cd.offset = offset
	cd.size = 0
	cd.read = 0
	return nil
}

// Key is used to get the aes key.
func (cd *CTRDecrypter) Key() []byte {
	return cd.ctr.Key()
}

// Image is used to get the original image in the under Reader.
func (cd *CTRDecrypter) Image() image.Image {
	return cd.reader.Image()
}

// Cap is used to calculate the capacity that can decrypt from the under Reader.
func (cd *CTRDecrypter) Cap() int64 {
	return cd.capacity
}

// Mode is used to get the mode about the under Reader.
func (cd *CTRDecrypter) Mode() Mode {
	return cd.reader.Mode()
}

// Algorithm is used to get the algorithm.
func (*CTRDecrypter) Algorithm() Algorithm {
	return AESWithCTR
}
