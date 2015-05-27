package core

import (
	"errors"
	"os"
	"os/signal"
	"sync"

	"github.com/jonog/redalert/notifiers"
)

type Service struct {
	checks    []*Check
	notifiers map[string]notifiers.Notifier
	wg        sync.WaitGroup
}

func NewService() *Service {
	return &Service{
		notifiers: make(map[string]notifiers.Notifier),
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

func (s *Service) RegisterCheck(check *Check, sendAlerts []string) error {
	err := check.AddNotifiers(s, sendAlerts)
	if err != nil {
		return err
	}
	s.checks = append(s.checks, check)
	return nil
}

func (s *Service) RegisterNotifier(notifier notifiers.Notifier) error {
	_, exists := s.notifiers[notifier.Name()]
	if exists {
		return errors.New("redalert: notifier already existing on service. name: " + notifier.Name())
	}
	s.notifiers[notifier.Name()] = notifier
	return nil
}
