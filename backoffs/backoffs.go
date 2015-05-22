package backoffs

import (
	"time"
)

type Backoff interface {
	Init() time.Duration
	Next(failCount int) time.Duration
}

type BackoffConfig struct {
	Type       string `json:"type"`
	Interval   int    `json:"interval"`
	Multiplier int    `json:"multiplier,omitempty"`
}

const (
	BackoffTypeConstant    = "constant"
	BackoffTypeLinear      = "linear"
	BackoffTypeExponential = "exponential"
)

// Creates new Backoff instance based on provided configuration
func BackoffFactory(cfg BackoffConfig) Backoff {
	switch cfg.Type {
	case BackoffTypeConstant:
		return NewConstantBackoff(cfg.Interval)
	case BackoffTypeLinear:
		return NewLinearBackoff(cfg.Interval)
	case BackoffTypeExponential:
		return NewExponentialBackoff(cfg.Interval, cfg.Multiplier)
	default:
		return NewConstantBackoff(cfg.Interval)
	}
}
