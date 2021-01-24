package lsb

import (
	"image"
	"image/png"
	"io"

	"github.com/pkg/errors"
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

var _ Writer = new(PNGWriter)

// PNGWriter implemented lsb Writer interface.
type PNGWriter struct {
	*pngCommon
}

// NewPNGWriter is used to create a png lsb writer.
func NewPNGWriter(img image.Image, mode Mode) (*PNGWriter, error) {
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

var _ Reader = new(PNGReader)

// PNGReader implemented lsb Reader interface.
type PNGReader struct {
	*pngCommon
}

// NewPNGReader is used to create a png lsb reader.
func NewPNGReader(r io.Reader) (*PNGReader, error) {
	p, err := png.Decode(r)
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
