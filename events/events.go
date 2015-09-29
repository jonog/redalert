package events

import (
	"strconv"
	"strings"
	"time"
)

const (
	redalert   = "redalert"
	greenalert = "greenalert"
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
	return e.HasTag(redalert)
}

func (e *Event) MarkRedAlert() {
	e.AddTag(redalert)
}

func (e *Event) IsGreenAlert() bool {
	return e.HasTag(greenalert)
}

func (e *Event) MarkGreenAlert() {
	e.AddTag(greenalert)
}

func (e *Event) HasTag(t string) bool {
	_, exists := e.Tags[t]
	return exists
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
