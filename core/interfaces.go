package core

import "github.com/jonog/redalert/notifiers"

type Notifier interface {
	Notify(notifiers.Message) error
	Name() string
}

type EventStorage interface {
	Store(*Event) error
	Last() (*Event, error)
	GetRecent() ([]*Event, error)
}
