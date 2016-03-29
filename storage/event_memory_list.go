package storage

import (
	"container/list"

	"github.com/jonog/redalert/events"
)

type MemoryList struct {
	maxEvents  int
	lastEvent  *events.Event
	history    *list.List
	failCounts map[string]int
}

func NewMemoryList(capacity int) *MemoryList {
	return &MemoryList{capacity, nil, list.New(), make(map[string]int)}
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

	// if no events, return empty array
	if len(es) == 0 {
		return make([]*events.Event, 0), nil
	}

	return es, nil
}

func (l *MemoryList) IncrFailCount(trigger string) (int, error) {
	l.failCounts[trigger]++
	return l.failCounts[trigger], nil
}

func (l *MemoryList) ResetFailCount(trigger string) error {
	l.failCounts[trigger] = 0
	return nil
}
