package core

import (
	"github.com/jonog/redalert/checks"
	"github.com/jonog/redalert/notifiers"
)

type Alert interface {
	Trigger(notifiers.Message) error
	Name() string
}

type Checker interface {
	Check() (map[string]float64, error)
	MetricInfo(string) checks.MetricInfo
	RedAlertMessage() string
	GreenAlertMessage() string
}
