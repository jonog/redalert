package events

import (
	"strconv"
	"time"
)

type Event struct {
	Time time.Time
	Type string
	Data map[string]*float64
}

func NewEvent(data map[string]*float64) *Event {
	return &Event{Time: time.Now(), Data: data}
}

func (e *Event) SetType(t string) {
	e.Type = t
}

func (e *Event) IsRedAlert() bool {
	return e.Type == "redalert"
}

func (e *Event) DisplayMetric(metric string) string {
	if e.Data[metric] == nil {
		return ""
	}
	return strconv.FormatFloat(*e.Data[metric], 'f', 1, 64)
}
