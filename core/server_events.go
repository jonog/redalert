package core

import (
	"strconv"
	"strings"
)

func (s *Server) StoreEvent(event *Event) {
	s.LastEvent = event
	s.EventHistory.PushFront(event)
	if s.EventHistory.Len() > MaxEventsStored {
		s.EventHistory.Remove(s.EventHistory.Back())
	}
}

func (s *Server) GetEvents() string {
	var output []string
	for e := s.EventHistory.Front(); e != nil; e = e.Next() {
		event := e.Value.(*Event)
		if event != nil {
			output = append([]string{strconv.FormatInt(event.Latency.Nanoseconds()/1e6, 10)}, output...)
		}
	}
	return strings.Join(output, ",")
}
