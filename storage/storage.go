package storage

import (
	"time"

	"github.com/jonog/redalert/checks"
	"github.com/jonog/redalert/events"
	"github.com/jonog/redalert/notifiers"
)

type EventStorage interface {
	Store(*events.Event) error
	Last() (*events.Event, error)
	GetRecent() ([]*events.Event, error)
}

type ConfigStorage interface {
	Notifications() ([]notifiers.Config, error)
	Checks() ([]checks.Config, error)
}

type Counter interface {
	Inc(label string, count int) int
	Reset(label string)
}

func NewBasicCounter() *BasicCounter {
	return &BasicCounter{counts: make(map[string]int)}
}

type BasicCounter struct {
	counts map[string]int
}

func (b *BasicCounter) Inc(label string, count int) int {
	b.counts[label] += count
	return b.counts[label]
}

func (b *BasicCounter) Reset(label string) {
	b.counts[label] = 0
	return
}

type Tracker interface {
	Track(label string)
}

func NewBasicTracker() *BasicTracker {
	return &BasicTracker{events: make(map[string]time.Time)}
}

type BasicTracker struct {
	events map[string]time.Time
}

func (b *BasicTracker) Track(label string) {
	b.events[label] = time.Now()
}
