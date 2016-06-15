package checks

import (
	"encoding/json"
	"errors"
	"log"

	"github.com/jonog/redalert/assertions"
	"github.com/jonog/redalert/backoffs"
	"github.com/jonog/redalert/data"
)

// The Checker implements a type of status check / mechanism of data collection
// which may be used for triggering alerts
type Checker interface {
	Check() (data.CheckResponse, error)
	MetricInfo(string) MetricInfo
	MessageContext() string
}

type MetricInfo struct {
	Unit string
}

/////////////////
// Initialisation
/////////////////

type Config struct {
	Name       string              `json:"name"`
	Type       string              `json:"type"`
	SendAlerts []string            `json:"send_alerts"`
	Backoff    backoffs.Config     `json:"backoff"`
	Config     json.RawMessage     `json:"config"`
	Assertions []assertions.Config `json:"assertions"`
}

var registry = make(map[string]func(Config, *log.Logger) (Checker, error))

func Register(name string, constructorFn func(Config, *log.Logger) (Checker, error)) {
	registry[name] = constructorFn
}

func New(config Config, logger *log.Logger) (Checker, error) {
	checkerFn, ok := registry[config.Type]
	if !ok {
		return nil, errors.New("checks: checker unavailable: " + config.Type)
	}
	return checkerFn(config, logger)
}
