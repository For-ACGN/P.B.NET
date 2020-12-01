// +build go1.10, !go1.12

package strings

// Cap returns the capacity of the builder's underlying byte slice. It is the
// total space allocated for the string being built and includes any bytes
// already written.
//
// From go1.12
func (b *Builder) Cap() int { return cap(b.buf) }
