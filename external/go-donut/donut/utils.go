package donut

import (
	"bytes"
	"crypto/rand"
	"io"
	"math/big"
	"net/http"
)

// RandomString - generates random string of given length.
func RandomString(size int) string {
	b := make([]byte, size)
	for i := 0; i < size; i++ {
		r, _ := rand.Int(rand.Reader, big.NewInt(25))
		b[i] = 97 + byte(r.Int64()) // a=97
	}
	return string(b)
}

// RandomBytes : Generates as many random bytes as you ask for, returns them as []byte.
func RandomBytes(count int) ([]byte, error) {
	b := make([]byte, count)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return nil, err
	}
	return b, nil
}

// DownloadFile will download an URL to a byte buffer
func DownloadFile(url string) (*bytes.Buffer, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	buf := bytes.NewBuffer([]byte{})
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		return nil, err
	}
	return buf, nil
}