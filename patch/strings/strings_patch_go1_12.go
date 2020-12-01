// +build go1.10, !go1.12

package strings

// ReplaceAll returns a copy of the string s with all
// non-overlapping instances of old replaced by new.
// If old is empty, it matches at the beginning of the string
// and after each UTF-8 sequence, yielding up to k+1 replacements
// for a k-rune string.
//
// From go1.12
func ReplaceAll(s, old, new string) string {
	return Replace(s, old, new, -1)
}
