package events

import (
	"strconv"
	"strings"
	"time"
)

type Event struct {
	Time time.Time
	Data map[string]*float64
	// TODO: change to map[string]struct{}
	Tags []string
}

func NewEvent(data map[string]*float64) *Event {
	return &Event{Time: time.Now(), Data: data, Tags: []string{}}
}

func (e *Event) AddTag(t string) {
	e.Tags = append(e.Tags, t)
}

func (e *Event) IsRedAlert() bool {
	for _, tag := range e.Tags {
		if tag == "redalert" {
			return true
		}
	}
	return false
}

func (e *Event) HasTag(t string) bool {
	for _, tag := range e.Tags {
		if tag == t {
			return true
		}
	}
	return false
}

func (e *Event) DisplayMetric(metric string) string {
	if e.Data[metric] == nil {
		return ""
	}
	return strconv.FormatFloat(*e.Data[metric], 'f', 1, 64)
}

func (e *Event) DisplayTags() string {
	return strings.Join(e.Tags, " ")
}
