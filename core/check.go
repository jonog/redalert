package core

import (
	"container/list"
	"errors"
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

func (s *Service) RegisterCheck(config CheckConfig) error {
	check, err := NewCheck(config)
	if err != nil {
		return err
	}
	check.service = s
	err = check.AddAlerts(config.Alerts)
	if err != nil {
		return err
	}
	s.checks = append(s.checks, check)
	return nil
}

func NewCheck(config CheckConfig) (*Check, error) {

	logger := log.New(os.Stdout, config.Name+" ", log.Ldate|log.Ltime)

	var checker Checker
	switch config.Type {
	case "web-ping":
		checker = checks.NewWebPinger(config.Address, logger)
	case "scollector":
		checker = checks.NewSCollector(config.Host)
	default:
		return nil, errors.New("redalert: unknown notifier")
	}

	return &Check{
		Name:         config.Name,
		Interval:     config.Interval,
		Alerts:       make([]Notifier, 0),
		Log:          logger,
		EventHistory: list.New(),
		Checker:      checker,
	}, nil
}

func (c *Check) AddAlerts(names []string) error {
	for _, name := range names {
		alert, err := getAlert(c.service, name)
		if err != nil {
			return err
		}
		c.Alerts = append(c.Alerts, alert)
	}
	return nil
}

func getAlert(service *Service, name string) (Notifier, error) {
	alert, ok := service.Notifiers[name]
	if !ok {
		return nil, errors.New("redalert: notifier requested has not be registered. name: " + name)
	}
	return alert, nil
}

func (c *Check) incrFailCount() {
	c.failCount++
}

func (c *Check) resetFailCount() {
	c.failCount = 0
}
