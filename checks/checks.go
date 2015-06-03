package checks

import (
	"errors"
	"log"
	"regexp"
	"strconv"

	"github.com/jonog/redalert/backoffs"
)

// The Checker implements a type of status check / mechanism of data collection
// which may be used for triggering alerts
type Checker interface {
	Check() (Metrics, error)
	MetricInfo(string) MetricInfo
	MessageContext() string
}

type MetricInfo struct {
	Unit string
}

type Metrics map[string]*float64

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

	Triggers []Trigger `json:"triggers"`
}

type Trigger struct {
	Metric   string `json:"metric"`
	Criteria string `json:"criteria"`
}

func (t Trigger) MeetsCriteria(data Metrics) bool {
	currentVal, exists := data[t.Metric]
	if !exists || currentVal == nil {
		return false
	}
	re := regexp.MustCompile(`([(>)(<)]?=?)([\d\.]+)`)
	matches := re.FindStringSubmatch(t.Criteria)
	if len(matches) != 3 {
		return false
	}
	operator := matches[1]
	criteriaValStr := matches[2]
	criteriaVal, err := strconv.ParseFloat(criteriaValStr, 64)
	if err != nil {
		return false
	}
	return evaluate(operator, *currentVal, criteriaVal)
}

func evaluate(operator string, num1, num2 float64) bool {
	if operator == ">" && num1 > num2 {
		return true
	} else if operator == ">=" && num1 >= num2 {
		return true
	} else if operator == "<" && num1 < num2 {
		return true
	} else if operator == "<=" && num1 <= num2 {
		return true
	} else if (operator == "==" || operator == "=") && num1 == num2 {
		return true
	} else {
		return false
	}
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
