package backoffs

import (
	"time"
)

type ExponentialBackoff struct {
	Interval   int
	Multiplier int
}

func NewExponentialBackoff(interval, multiplier int) *ExponentialBackoff {
	b := new(ExponentialBackoff)
	b.Interval = interval

	if multiplier < 0 {
		multiplier = 1
	}

	b.Multiplier = multiplier
	return b
}

func (b *ExponentialBackoff) Init() time.Duration {
	return time.Second * time.Duration(b.Interval)
}

func (b *ExponentialBackoff) Next(failCount int) time.Duration {
	return time.Second * time.Duration(failCount*b.Interval*b.Multiplier)
}
