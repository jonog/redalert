package core

import (
	"os"
	"os/signal"
	"sync"

	"github.com/jonog/redalert/checks"
	"github.com/jonog/redalert/notifiers"
)

type Service struct {
	checks    []*Check
	Notifiers map[string]notifiers.Notifier
	wg        sync.WaitGroup
}

func NewService() *Service {
	return &Service{
		Notifiers: make(map[string]notifiers.Notifier),
	}
}

// Start starts the monitoring system, by starting each check.
// The service runs until a signal is received to stop the service.
func (s *Service) Start() {

	// use this to keep the service running, even if no monitoring is occuring
	s.wg.Add(1)

	for _, check := range s.checks {
		go check.Start()
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for _ = range c {
			s.wg.Done()
		}
	}()

}

// KeepRunning is called to wait indefinitely while the service runs.
func (s *Service) KeepRunning() {
	s.wg.Wait()
}

func (s *Service) Checks() []*Check {
	return s.checks
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

func (s *Service) RegisterNotifier(config notifiers.Config) error {
	notifier, err := notifiers.New(config)
	if err != nil {
		return err
	}
	s.Notifiers[notifier.Name()] = notifier
	return nil
}
