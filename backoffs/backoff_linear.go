package backoffs

import (
	"time"
)

type LinearBackoff struct {
	Interval int
}

func NewLinearBackoff(interval int) *LinearBackoff {
	b := new(LinearBackoff)
	b.Interval = interval
	return b
}

func (b *LinearBackoff) Init() time.Duration {
	return time.Second * time.Duration(b.Interval)
}

func (b *LinearBackoff) Next(failCount int) time.Duration {
	return time.Second * time.Duration(failCount*b.Interval)
}
