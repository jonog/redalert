package core

import (
	"container/list"
	"log"
	"os"

	"github.com/jonog/redalert/checks"
)

type CheckConfig struct {
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	Interval int      `json:"interval"`
	Alerts   []string `json:"alerts"`

	// used for web-ping
	Address string `json:"address"`

	// used for scollector
	Host string `json:"host"`
}

type Check struct {
	Name         string
	Type         string // e.g. future options: web-ping, ssh-ping, query
	Interval     int
	Alerts       []Notifier
	Log          *log.Logger
	service      *Service
	failCount    int
	LastEvent    *Event
	EventHistory *list.List
	Checker      Checker
}

func NewCheck(config CheckConfig) *Check {

	logger := log.New(os.Stdout, config.Name+" ", log.Ldate|log.Ltime)

	var checker Checker
	switch config.Type {
	case "web-ping":
		checker = checks.NewWebPinger(config.Address, logger)
	case "scollector":
		checker = checks.NewSCollector(config.Host)
	default:
		panic("unknown check type")
	}

	return &Check{
		Name:         config.Name,
		Interval:     config.Interval,
		Alerts:       make([]Notifier, 0),
		Log:          logger,
		EventHistory: list.New(),
		Checker:      checker,
	}
}

func (c *Check) AddAlerts(names []string) {
	for _, name := range names {
		c.Alerts = append(c.Alerts, getAlert(c.service, name))
	}
}

func getAlert(service *Service, name string) Notifier {
	alert, ok := service.Notifiers[name]
	if !ok {
		panic("Notifier has not been registered!")
	}
	return alert
}

func (c *Check) incrFailCount() {
	c.failCount++
}

func (c *Check) resetFailCount() {
	c.failCount = 0
}
