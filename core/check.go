package core

import (
	"container/list"
	"errors"
	"log"
	"os"

	"github.com/jonog/redalert/checks"
)

type Check struct {
	Name     string
	Type     string // e.g. future options: web-ping, ssh-ping, query
	Interval int

	Notifiers []Notifier

	Log     *log.Logger
	service *Service

	failCount    int
	LastEvent    *Event
	EventHistory *list.List

	Checker Checker
}

func (s *Service) RegisterCheck(config checks.Config) error {
	check, err := NewCheck(config)
	if err != nil {
		return err
	}
	check.service = s
	err = check.AddNotifiers(config.SendAlerts)
	if err != nil {
		return err
	}
	s.checks = append(s.checks, check)
	return nil
}

func NewCheck(config checks.Config) (*Check, error) {

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
		Notifiers:    make([]Notifier, 0),
		Log:          logger,
		EventHistory: list.New(),
		Checker:      checker,
	}, nil
}

func (c *Check) AddNotifiers(names []string) error {
	for _, name := range names {
		notifier, err := getNotifier(c.service, name)
		if err != nil {
			return err
		}
		c.Notifiers = append(c.Notifiers, notifier)
	}
	return nil
}

func getNotifier(service *Service, name string) (Notifier, error) {
	notifier, ok := service.Notifiers[name]
	if !ok {
		return nil, errors.New("redalert: notifier requested has not be registered. name: " + name)
	}
	return notifier, nil
}

func (c *Check) incrFailCount() {
	c.failCount++
}

func (c *Check) resetFailCount() {
	c.failCount = 0
}
