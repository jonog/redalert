package backoffs

import (
	"time"
)

type ConstantBackoff struct {
	Interval int
}

func NewConstantBackoff(interval int) *ConstantBackoff {
	b := new(ConstantBackoff)
	b.Interval = interval
	return b
}

func (b *ConstantBackoff) Init() time.Duration {
	return time.Second * time.Duration(b.Interval)
}

func (b *ConstantBackoff) Next(failCount int) time.Duration {
	return time.Second * time.Duration(b.Interval)
}
