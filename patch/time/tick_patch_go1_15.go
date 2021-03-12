// +build go1.14,!go1.15

package time

// Reset is a padding function for pass build about file:
// script/install/patch/verify.go
func (t *Ticker) Reset(d Duration) {
	panic("time: invalid call Ticker.Reset")
}
