package core

import (
	"errors"
	"log"
	"os"

	"github.com/jonog/redalert/checks"
	"github.com/jonog/redalert/notifiers"
)

var MaxEventsStored = 100

type Check struct {
	Name     string
	Type     string // e.g. future options: web-ping, ssh-ping, query
	Interval int

	Notifiers []notifiers.Notifier

	Log *log.Logger

	failCount int

	service *Service
	Store   EventStorage

	Checker checks.Checker
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

	checker, err := checks.New(config, logger)
	if err != nil {
		return nil, err
	}

	return &Check{
		Name:      config.Name,
		Interval:  config.Interval,
		Notifiers: make([]notifiers.Notifier, 0),
		Log:       logger,
		Store:     NewMemoryList(MaxEventsStored),
		Checker:   checker,
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

func getNotifier(service *Service, name string) (notifiers.Notifier, error) {
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
