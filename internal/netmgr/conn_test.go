package netmgr

import (
	"testing"
	"time"

	"golang.org/x/time/rate"
)

func TestRate(t *testing.T) {
	rate.NewLimiter(rate.Every(time.Second), 0)
}
