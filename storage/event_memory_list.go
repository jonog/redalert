package storage

import (
	"container/list"
	"sync"

	"github.com/jonog/redalert/events"
)

type MemoryList struct {
	mu        sync.RWMutex
	maxEvents int
	lastEvent *events.Event
	history   *list.List
}

func NewMemoryList(capacity int) *MemoryList {
	return &MemoryList{maxEvents: capacity, history: list.New()}
}

func (l *MemoryList) Store(event *events.Event) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.lastEvent = event
	l.history.PushFront(event)
	if l.history.Len() > l.maxEvents {
		l.history.Remove(l.history.Back())
	}
	return nil
}

func (l *MemoryList) Last() (*events.Event, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.lastEvent, nil
}

func (l *MemoryList) GetRecent() ([]*events.Event, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	var es []*events.Event
	for e := l.history.Front(); e != nil; e = e.Next() {
		event := e.Value.(*events.Event)
		if event != nil {
			es = append(es, event)
		}
	}

	// if no events, return empty array
	if len(es) == 0 {
		return make([]*events.Event, 0), nil
	}

	return es, nil
}
