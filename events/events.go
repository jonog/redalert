package events

import (
	"strconv"
	"strings"
	"time"
)

type Event struct {
	Time    RFCTime             `json:"time"`
	Metrics map[string]*float64 `json:"data"`
	Tags    map[string]string   `json:"tags"`
}

func NewEvent(metrics map[string]*float64) *Event {
	return &Event{
		Time:    RFCTime{time.Now()},
		Metrics: metrics,
		Tags:    make(map[string]string),
	}
}

func (e *Event) AddTag(t string) {
	e.Tags[t] = ""
}

func (e *Event) IsRedAlert() bool {
	_, exists := e.Tags["redalert"]
	return exists
}

func (e *Event) HasTag(t string) bool {
	for tag := range e.Tags {
		if tag == t {
			return true
		}
	}
	return false
}

func (e *Event) DisplayMetric(metric string) string {
	if e.Metrics[metric] == nil {
		return ""
	}
	return strconv.FormatFloat(*e.Metrics[metric], 'f', 1, 64)
}

func (e *Event) DisplayTags() string {
	// required as used in template
	if e == nil {
		return ""
	}
	var keys []string
	for tag := range e.Tags {
		keys = append(keys, tag)
	}
	return strings.Join(keys, " ")
}
