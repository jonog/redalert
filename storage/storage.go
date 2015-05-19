package storage

import "github.com/jonog/redalert/events"

type EventStorage interface {
	Store(*events.Event) error
	Last() (*events.Event, error)
	GetRecent() ([]*events.Event, error)
}
