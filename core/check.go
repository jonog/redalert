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
	Address  string   `json:"address"`
	Interval int      `json:"interval"`
	Alerts   []string `json:"alerts"`
}

type Check struct {
	Name         string
	Type         string // e.g. future options: web-ping, ssh-ping, query
	Interval     int
	Alerts       []Alert
	Log          *log.Logger
	service      *Service
	failCount    int
	LastEvent    *Event
	EventHistory *list.List
	Checker      Checker
}

func NewCheck(config CheckConfig) *Check {

	var checker Checker
	if config.Type == "web-ping" {
		id := config.Name + " (" + config.Address + ")"
		checker = checks.NewWebPinger(id, config.Address)
	} else {
		// todo - add validation, handle err gracefully
		panic("unknown type")
	}

	return &Check{
		Name:         config.Name,
		Interval:     config.Interval,
		Alerts:       make([]Alert, 0),
		Log:          log.New(os.Stdout, config.Name+" ", log.Ldate|log.Ltime),
		EventHistory: list.New(),
		Checker:      checker,
	}
}

func (c *Check) AddAlerts(names []string) {
	for _, name := range names {
		c.Alerts = append(c.Alerts, getAlert(c.service, name))
	}
}

func getAlert(service *Service, name string) Alert {
	alert, ok := service.Alerts[name]
	if !ok {
		panic("Alert has not been registered!")
	}
	return alert
}

func (c *Check) incrFailCount() {
	c.failCount++
}

func (c *Check) resetFailCount() {
	c.failCount = 0
}
