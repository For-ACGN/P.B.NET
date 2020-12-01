// +build go1.10, !go1.14

package http

import (
	"net/textproto"
)

// Values returns all values associated with the given key.
// It is case insensitive; textproto.CanonicalMIMEHeaderKey is
// used to canonicalize the provided key. To use non-canonical
// keys, access the map directly.
// The returned slice is not a copy.
//
// From go1.14
func (h Header) Values(key string) []string {
	return textproto.MIMEHeader(h).Values(key)
}
