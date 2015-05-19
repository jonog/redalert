package core

import "container/list"

type EventStorage interface {
	Store(*Event) error
	Last() (*Event, error)
	GetRecent() ([]*Event, error)
}

type MemoryList struct {
	maxEvents int
	lastEvent *Event
	history   *list.List
}

func NewMemoryList(capacity int) *MemoryList {
	return &MemoryList{capacity, nil, list.New()}
}

func (l *MemoryList) Store(event *Event) error {
	l.lastEvent = event
	l.history.PushFront(event)
	if l.history.Len() > l.maxEvents {
		l.history.Remove(l.history.Back())
	}
	return nil
}

func (l *MemoryList) Last() (*Event, error) {
	return l.lastEvent, nil
}

func (l *MemoryList) GetRecent() ([]*Event, error) {
	var events []*Event
	for e := l.history.Front(); e != nil; e = e.Next() {
		event := e.Value.(*Event)
		if event != nil {
			events = append(events, event)
		}
	}
	return events, nil
}
