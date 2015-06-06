package storage

import (
	"container/list"

	"github.com/jonog/redalert/events"
)

type MemoryList struct {
	maxEvents int
	lastEvent *events.Event
	history   *list.List
}

func NewMemoryList(capacity int) *MemoryList {
	return &MemoryList{capacity, nil, list.New()}
}

func (l *MemoryList) Store(event *events.Event) error {
	l.lastEvent = event
	l.history.PushFront(event)
	if l.history.Len() > l.maxEvents {
		l.history.Remove(l.history.Back())
	}
	return nil
}

func (l *MemoryList) Last() (*events.Event, error) {
	return l.lastEvent, nil
}

func (l *MemoryList) GetRecent() ([]*events.Event, error) {
	var es []*events.Event
	for e := l.history.Front(); e != nil; e = e.Next() {
		event := e.Value.(*events.Event)
		if event != nil {
			es = append(es, event)
		}
	}
	return es, nil
}
