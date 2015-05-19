package core

import (
	"strconv"
	"strings"
	"time"
)

type Event struct {
	Check *Check
	Time  time.Time
	Type  string
	Data  map[string]float64
}

func NewRedAlert(check *Check, data map[string]float64) *Event {
	return &Event{Check: check, Time: time.Now(), Type: "redalert", Data: data}
}

func NewGreenAlert(check *Check, data map[string]float64) *Event {
	return &Event{Check: check, Time: time.Now(), Type: "greenalert", Data: data}
}

func (e *Event) isRedAlert() bool {
	return e.Type == "redalert"
}

func (e *Event) isGreenAlert() bool {
	return e.Type == "greenalert"
}

func (e *Event) ShortMessage() string {

	if e.isRedAlert() {
		return e.Check.Checker.RedAlertMessage()
	}

	if e.isGreenAlert() {
		return e.Check.Checker.GreenAlertMessage()
	}

	return ""
}

func (e *Event) DisplayMetric(metric string) string {
	return strconv.FormatFloat(e.Data[metric], 'f', 1, 64)
}

func (c *Check) RecentMetrics(metric string) string {
	events, err := c.Store.GetRecent()
	if err != nil {
		c.Log.Println("ERROR: retrieving recent events")
		return ""
	}
	var output []string
	for _, event := range events {
		output = append([]string{event.DisplayMetric(metric)}, output...)
	}
	return strings.Join(output, ",")
}
