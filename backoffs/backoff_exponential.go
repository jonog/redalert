package backoffs

import (
	"time"
)

type ExponentialBackoff struct {
	Interval   int
	Multiplier int
}

// Exponential backoff multiplies initial interval by failed request count and specified multiplier
func NewExponentialBackoff(interval, multiplier int) *ExponentialBackoff {
	b := new(ExponentialBackoff)
	b.Interval = interval

	if multiplier <= 0 {
		multiplier = 1
	}

	b.Multiplier = multiplier
	return b
}

// Returns initial interval
func (b *ExponentialBackoff) Init() time.Duration {
	return time.Second * time.Duration(b.Interval)
}

// Returns next interval based on failed requests count
func (b *ExponentialBackoff) Next(failCount int) time.Duration {
	return time.Second * time.Duration(failCount*b.Interval*b.Multiplier)
}
