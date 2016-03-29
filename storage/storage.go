package storage

import (
	"github.com/jonog/redalert/checks"
	"github.com/jonog/redalert/events"
	"github.com/jonog/redalert/notifiers"
)

type EventStorage interface {
	Store(*events.Event) error
	Last() (*events.Event, error)
	GetRecent() ([]*events.Event, error)
	IncrFailCount(trigger string) (int, error)
	ResetFailCount(trigger string) error
}

type ConfigStorage interface {
	Notifications() ([]notifiers.Config, error)
	Checks() ([]checks.Config, error)
}
