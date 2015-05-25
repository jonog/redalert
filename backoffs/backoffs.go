package backoffs

import (
	"time"
)

const DefaultInterval = 10

type Backoff interface {
	Init() time.Duration
	Next(failCount int) time.Duration
}

type Config struct {
	Type       string `json:"type"`
	Interval   *int   `json:"interval"`
	Multiplier int    `json:"multiplier,omitempty"`
}

const (
	TypeConstant    = "constant"
	TypeLinear      = "linear"
	TypeExponential = "exponential"
)

// Creates new Backoff instance based on provided configuration
func New(cfg Config) Backoff {

	// if no interval is provided (nil), set to default value
	var interval int
	if cfg.Interval == nil {
		interval = DefaultInterval
	} else {
		interval = *cfg.Interval
	}

	switch cfg.Type {
	case TypeConstant:
		return NewConstant(interval)
	case TypeLinear:
		return NewLinear(interval)
	case TypeExponential:
		return NewExponential(interval, cfg.Multiplier)
	default:
		return NewConstant(interval)
	}
}
