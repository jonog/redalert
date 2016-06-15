package events

import (
	"strconv"
	"strings"
	"time"

	"github.com/jonog/redalert/data"
)

const (
	redalert   = "redalert"
	greenalert = "greenalert"
)

type Event struct {
	Time     RFCTime            `json:"time"`
	Data     data.CheckResponse `json:"data"`
	Tags     map[string]string  `json:"tags"`
	Messages []string           `json:"messages"`
}

func NewEvent(checkData data.CheckResponse) *Event {
	return &Event{
		Time:     RFCTime{time.Now()},
		Data:     checkData,
		Tags:     make(map[string]string),
		Messages: []string{},
	}
}

func (e *Event) AddTag(t string) {
	e.Tags[t] = ""
}

func (e *Event) IsRedAlert() bool {
	return e.HasTag(redalert)
}

func (e *Event) MarkRedAlert(messages []string) {
	e.Messages = messages
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
	if e.Data.Metrics[metric] == nil {
		return ""
	}
	return strconv.FormatFloat(*e.Data.Metrics[metric], 'f', 1, 64)
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
