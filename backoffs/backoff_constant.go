package backoffs

import (
	"time"
)

type Constant struct {
	Interval int
}

// Constant backoff has always the same interval
func NewConstant(interval int) *Constant {
	b := new(Constant)
	b.Interval = interval
	return b
}

// Returns initial interval
func (b *Constant) Init() time.Duration {
	return time.Second * time.Duration(b.Interval)
}

// Returns next interval based on failed requests count
func (b *Constant) Next(failCount int) time.Duration {
	return time.Second * time.Duration(b.Interval)
}
