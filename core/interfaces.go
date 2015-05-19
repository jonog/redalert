package core

import (
	"github.com/jonog/redalert/checks"
	"github.com/jonog/redalert/notifiers"
)

type Notifier interface {
	Notify(notifiers.Message) error
	Name() string
}

// The Checker implements a type of status check / mechanism of data collection
// which may be used for triggering alerts
type Checker interface {
	Check() (checks.Metrics, error)
	MetricInfo(string) checks.MetricInfo
	RedAlertMessage() string
	GreenAlertMessage() string
}

type EventStorage interface {
	Store(*Event) error
	Last() (*Event, error)
	GetRecent() ([]*Event, error)
}
