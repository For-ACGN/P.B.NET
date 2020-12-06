// +build go1.10,!go1.14

package json

// InputOffset returns the input stream byte offset of the current decoder position.
// The offset gives the location of the end of the most recently returned token
// and the beginning of the next token.
//
// From go1.14
func (dec *Decoder) InputOffset() int64 {
	return dec.scanned + int64(dec.scanp)
}
