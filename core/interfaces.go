package core

import "github.com/jonog/redalert/checks"

type Alert interface {
	Trigger(*Event) error
	Name() string
}

type Checker interface {
	Check() (map[string]float64, error)
	MetricInfo(string) checks.MetricInfo
	RedAlertMessage() string
	GreenAlertMessage() string
}
