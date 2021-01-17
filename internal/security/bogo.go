package security

import (
	"context"
	"time"

	"project/internal/random"
)

// Bogo is used to use bogo sort for wait time, If timeout, it will interrupt.
type Bogo struct {
	number  int             // random number count
	timeout time.Duration   // wait timeout
	results map[string]bool // confuse result
	key     string          // key for store result

	ctx    context.Context
	cancel context.CancelFunc
}

// NewBogo is used to create a bogo waiter, if number is too large,
// Bogo.Wait() will block a lot of time.
func NewBogo(number int, timeout time.Duration) *Bogo {
	if number < 2 {
		number = 2
	}
	if timeout < 1 || timeout > 5*time.Minute {
		timeout = 10 * time.Second
	}
	rand := random.NewRand()
	b := Bogo{
		number:  number,
		timeout: timeout,
		results: make(map[string]bool),
		key:     rand.String(32 + rand.Intn(32)),
	}
	b.ctx, b.cancel = context.WithTimeout(context.Background(), b.timeout)
	return &b
}

// Wait is used to wait bogo sort.
func (bogo *Bogo) Wait() {
	defer bogo.cancel()
	rand := random.NewRand()
	// generate random number
	num := make([]int, bogo.number)
	for i := 0; i < bogo.number; i++ {
		num[i] = rand.Intn(100000)
	}
	// confuse result map
	// max = 256 * 64 = 16 KB
	c := 128 + rand.Intn(128)
	for i := 0; i < c; i++ {
		result := rand.String(32 + rand.Intn(32))
		bogo.results[result] = true
	}
swap:
	for {
		// check timeout
		select {
		case <-bogo.ctx.Done():
			return
		default:
		}
		// swap
		for i := 0; i < bogo.number; i++ {
			j := rand.Intn(bogo.number)
			num[i], num[j] = num[j], num[i]
		}
		// check is sorted
		for i := 1; i < bogo.number; i++ {
			if num[i-1] > num[i] {
				continue swap
			}
		}
		// set result
		bogo.results[bogo.key] = true
		return
	}
}

// Compare is used to compare the result is correct.
func (bogo *Bogo) Compare() bool {
	return bogo.results[bogo.key]
}

// Stop is used ot stop wait.
func (bogo *Bogo) Stop() {
	bogo.cancel()
}
