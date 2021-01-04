package random

import (
	"sync"
	"time"
)

// MaxSleepTime is used to prevent sleep dead.
const MaxSleepTime = 30 * time.Minute

// Sleeper contain a timer and rand for reuse.
// sleep total time = fixed + [0, random)
type Sleeper struct {
	timer *time.Timer
	rand  *Rand
	once  sync.Once
}

// NewSleeper is used to create a sleeper.
func NewSleeper() *Sleeper {
	timer := time.NewTimer(time.Minute)
	timer.Stop()
	return &Sleeper{timer: timer, rand: NewRand()}
}

// SleepSecond is used to sleep random second.
func (s *Sleeper) SleepSecond(fixed, random uint) <-chan time.Time {
	return s.SleepMillisecond(fixed*1000, random*1000)
}

// SleepMillisecond is used to sleep random millisecond.
func (s *Sleeper) SleepMillisecond(fixed, random uint) <-chan time.Time {
	d := s.calculateDuration(fixed, random)
	s.timer.Reset(d)
	select {
	case <-s.timer.C:
	default:
	}
	return s.timer.C
}

// calculateDuration is used to calculate actual duration.
func (s *Sleeper) calculateDuration(fixed, random uint) time.Duration {
	if fixed+random < 1 {
		fixed = 1000
	}
	random = uint(s.rand.Int(int(random)))
	total := time.Duration(fixed+random) * time.Millisecond
	if total > MaxSleepTime {
		total = MaxSleepTime
	}
	return total
}

// Stop is used to stop timer in sleeper.
func (s *Sleeper) Stop() {
	s.once.Do(func() { s.timer.Stop() })
}

// done, sleeper := random.SleepSecond(1, 1)
// defer sleeper.Stop()
// select {
// case <-done:
// case <-ctx.Done():
//     return ctx.Err()
// }
// ...

// SleepSecond is used to sleep random second.
func SleepSecond(fixed, random uint) (<-chan time.Time, *Sleeper) {
	sleeper := NewSleeper()
	return sleeper.SleepSecond(fixed, random), sleeper
}

// SleepMillisecond is used to sleep random millisecond.
func SleepMillisecond(fixed, random uint) (<-chan time.Time, *Sleeper) {
	sleeper := NewSleeper()
	return sleeper.SleepMillisecond(fixed, random), sleeper
}

// Sleep is used to sleep random second, it will not longer than 3 minutes.
func Sleep(fixed, random uint) {
	done, sleeper := SleepSecond(fixed, random)
	defer sleeper.Stop()
	timer := time.NewTimer(3 * time.Minute)
	defer timer.Stop()
	select {
	case <-done:
	case <-timer.C:
	}
}
