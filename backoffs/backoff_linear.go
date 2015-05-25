package backoffs

import (
	"time"
)

type Linear struct {
	Interval int
}

// Linear backoff multiplies initial interval by failed request count
func NewLinear(interval int) *Linear {
	b := new(Linear)
	b.Interval = interval
	return b
}

// Returns initial interval
func (b *Linear) Init() time.Duration {
	return time.Second * time.Duration(b.Interval)
}

// Returns next interval based on failed requests count
func (b *Linear) Next(failCount int) time.Duration {
	return time.Second * time.Duration(failCount*b.Interval)
}
