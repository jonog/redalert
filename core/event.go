package core

import (
	"strings"
	"time"
)

type Event struct {
	Server  *Server
	Time    time.Time
	Type    string
	Latency time.Duration
}

func NewRedAlert(server *Server, latency time.Duration) *Event {
	return &Event{Server: server, Time: time.Now(), Type: "redalert", Latency: latency}
}

func NewGreenAlert(server *Server, latency time.Duration) *Event {
	return &Event{Server: server, Time: time.Now(), Type: "greenalert", Latency: latency}
}

func (e *Event) isRedAlert() bool {
	return e.Type == "redalert"
}

func (e *Event) isGreenAlert() bool {
	return e.Type == "greenalert"
}

func (e *Event) ShortMessage() string {

	if e.isRedAlert() {
		return strings.Join([]string{"Uhoh,", e.Server.Name, "not responding. Failed ping to", e.Server.Address}, " ")
	}

	if e.isGreenAlert() {
		return strings.Join([]string{"Woo-hoo,", e.Server.Name, "is now reachable. Successful ping to", e.Server.Address}, " ")
	}

	return ""
}

func (e *Event) PrintLatency() int64 {
	return e.Latency.Nanoseconds() / 1e6
}
