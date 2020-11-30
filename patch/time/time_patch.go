// +build go1.10, !go1.13

package time

// Microseconds returns the duration as an integer microsecond count.
//
// From go1.13
func (d Duration) Microseconds() int64 { return int64(d) / 1e3 }

// Milliseconds returns the duration as an integer millisecond count.
//
// From go1.13
func (d Duration) Milliseconds() int64 { return int64(d) / 1e6 }
