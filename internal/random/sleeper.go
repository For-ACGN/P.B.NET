package random

import (
	"sync"
	"time"
)

// MaxSleepTime is used to prevent sleep dead.
const MaxSleepTime = 30 * time.Minute

// Sleeper contain a timer and rand for reuse.
// It is not multi goroutine safe except Stop.
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

// Sleep is used to sleep random second, it will block until time up.
// It will not longer than 3 minutes.
func (s *Sleeper) Sleep(fixed, random uint) {
	done := s.SleepSecond(fixed, random)
	timer := time.NewTimer(3 * time.Minute)
	defer timer.Stop()
	select {
	case <-done:
	case <-timer.C:
	}
}

// SleepSecond is used to sleep random second.
func (s *Sleeper) SleepSecond(fixed, random uint) <-chan time.Time {
	return s.SleepMillisecond(fixed*1000, random*1000)
}

// SleepMillisecond is used to sleep random millisecond.
func (s *Sleeper) SleepMillisecond(fixed, random uint) <-chan time.Time {
	s.timer.Reset(s.calculateTime(fixed, random))
	select {
	case <-s.timer.C:
	default:
	}
	return s.timer.C
}

// calculateTime is used to calculate actual time duration that need sleep.
func (s *Sleeper) calculateTime(fixed, random uint) time.Duration {
	if fixed+random < 1 {
		fixed = 1000
	}
	random = s.rand.Uintn(random)
	total := time.Duration(fixed+random) * time.Millisecond
	if total > MaxSleepTime {
		total = MaxSleepTime
	}
	return total
}

// Stop is used to stop timer in sleeper.
func (s *Sleeper) Stop() {
	s.once.Do(func() {
		s.timer.Stop()
	})
}

// Sleep is used to sleep random second, it will not longer than 3 minutes.
func Sleep(fixed, random uint) {
	sleeper := NewSleeper()
	defer sleeper.Stop()
	sleeper.Sleep(fixed, random)
}

// SleepMillisecond is used to sleep random millisecond, it will not longer than 3 minutes.
func SleepMillisecond(fixed, random uint) {
	sleeper := NewSleeper()
	defer sleeper.Stop()
	done := sleeper.SleepMillisecond(fixed, random)
	timer := time.NewTimer(3 * time.Minute)
	defer timer.Stop()
	select {
	case <-done:
	case <-timer.C:
	}
}
