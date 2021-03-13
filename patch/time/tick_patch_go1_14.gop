// +build go1.10,!go1.14

package time

// [warning] go1.14 also not has this method, but it will panic on it

// Reset stops a ticker and resets its period to the specified duration.
// The next tick will arrive after the new period elapses.
func (t *Ticker) Reset(d Duration) {
	if t.r.f == nil {
		panic("time: Reset called on uninitialized Ticker")
	}
	arg := t.r.arg
	stopTimer(&t.r)
	t.r = runtimeTimer{
		when:   when(d),
		period: int64(d),
		f:      sendTime,
		arg:    arg,
	}
	startTimer(&t.r)
}
