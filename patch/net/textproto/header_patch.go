// +build go1.10, !go1.14

package textproto

// Values returns all values associated with the given key.
// It is case insensitive; CanonicalMIMEHeaderKey is
// used to canonicalize the provided key. To use non-canonical
// keys, access the map directly.
// The returned slice is not a copy.
//
// From go1.14
func (h MIMEHeader) Values(key string) []string {
	if h == nil {
		return nil
	}
	return h[CanonicalMIMEHeaderKey(key)]
}
