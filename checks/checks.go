package checks

import (
	"errors"
	"log"

	"github.com/jonog/redalert/backoffs"
)

// The Checker implements a type of status check / mechanism of data collection
// which may be used for triggering alerts
type Checker interface {
	Check() (Metrics, error)
	MetricInfo(string) MetricInfo
	RedAlertMessage() string
	GreenAlertMessage() string
}

type MetricInfo struct {
	Unit string
}

type Metrics map[string]float64

/////////////////
// Initialisation
/////////////////

type Config struct {
	Name       string          `json:"name"`
	Type       string          `json:"type"`
	SendAlerts []string        `json:"send_alerts"`
	Backoff    backoffs.Config `json:"backoff"`

	// used for web-ping
	Address string `json:"address"`

	// used for scollector
	Host string `json:"host"`
}

var registry = make(map[string]func(Config, *log.Logger) (Checker, error))

func registerChecker(name string, constructorFn func(Config, *log.Logger) (Checker, error)) {
	registry[name] = constructorFn
}

func New(config Config, logger *log.Logger) (Checker, error) {
	checkerFn, ok := registry[config.Type]
	if !ok {
		return nil, errors.New("checks: checker unavailable: " + config.Type)
	}
	return checkerFn(config, logger)
}
