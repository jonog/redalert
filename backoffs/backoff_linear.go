package backoffs

import (
	"time"
)

type LinearBackoff struct {
	Interval int
}

// Linear backoff multiplies initial interval by failed request count
func NewLinearBackoff(interval int) *LinearBackoff {
	b := new(LinearBackoff)
	b.Interval = interval
	return b
}

// Returns initial interval
func (b *LinearBackoff) Init() time.Duration {
	return time.Second * time.Duration(b.Interval)
}

// Returns next interval based on failed requests count
func (b *LinearBackoff) Next(failCount int) time.Duration {
	return time.Second * time.Duration(failCount*b.Interval)
}
