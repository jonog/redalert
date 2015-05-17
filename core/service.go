package core

import (
	"os"
	"os/signal"
	"sync"
)

type Service struct {
	checks    []*Check
	Notifiers map[string]Notifier
	wg        sync.WaitGroup
}

func NewService() *Service {
	return &Service{
		Notifiers: make(map[string]Notifier),
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
