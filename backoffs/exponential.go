package backoffs

import (
	"math"
	"time"
)

type Exponential struct {
	Interval   int
	Multiplier int
}

const DefaultMultiplier = 2

// Exponential backoff provides a backoff implementation where the next delay upon failure is a
// multiplier of the previous delay
func NewExponential(interval int, multiplier *int) *Exponential {
	b := new(Exponential)
	b.Interval = interval

	// if multiplier is not provided or invalid, set to default value
	if multiplier == nil || *multiplier < 1 {
		b.Multiplier = DefaultMultiplier
	} else {
		b.Multiplier = *multiplier
	}

	return b
}

// Returns initial interval
func (b *Exponential) Init() time.Duration {
	return time.Second * time.Duration(b.Interval)
}

// Returns next interval based on failed requests count
func (b *Exponential) Next(failCountInt int) time.Duration {
	interval := float64(b.Interval)
	multiplier := float64(b.Multiplier)
	failCount := float64(failCountInt)
	return time.Second * time.Duration(interval*math.Pow(multiplier, failCount))
}
