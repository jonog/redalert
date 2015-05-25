package backoffs

import (
	"time"
)

type ConstantBackoff struct {
	Interval int
}

// Constant backoff has always the same interval
func NewConstantBackoff(interval int) *ConstantBackoff {
	b := new(ConstantBackoff)
	b.Interval = interval
	return b
}

// Returns initial interval
func (b *ConstantBackoff) Init() time.Duration {
	return time.Second * time.Duration(b.Interval)
}

// Returns next interval based on failed requests count
func (b *ConstantBackoff) Next(failCount int) time.Duration {
	return time.Second * time.Duration(b.Interval)
}
