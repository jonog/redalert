package core

import "time"

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
