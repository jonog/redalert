package core

import (
	"errors"
	"os"
	"os/signal"
	"sort"
	"sync"

	"github.com/jonog/redalert/notifiers"
)

type Service struct {
	checks    map[string]*Check
	notifiers map[string]notifiers.Notifier
	wg        sync.WaitGroup
}

func NewService() *Service {
	return &Service{
		checks:    make(map[string]*Check),
		notifiers: make(map[string]notifiers.Notifier),
	}
}

// Start starts the monitoring system, by starting each check.
// The service runs until a signal is received to stop the service.
func (s *Service) Start() {

	// use this to keep the service running, even if no monitoring is occuring
	s.wg.Add(1)

	for _, check := range s.checks {
		if check.Data.Enabled {
			go check.Start()
		}
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
	var checksArr []*Check
	for id := range s.checks {
		checksArr = append(checksArr, s.checks[id])
	}
	sort.Sort(ChecksArr(checksArr))
	return checksArr
}

func (s *Service) CheckByID(id string) (*Check, error) {
	check, exists := s.checks[id]
	if !exists {
		return nil, errors.New("service: check does not exist")
	}
	return check, nil
}

type ChecksArr []*Check

func (a ChecksArr) Len() int           { return len(a) }
func (a ChecksArr) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ChecksArr) Less(i, j int) bool { return a[i].ConfigRank < a[j].ConfigRank }

func (s *Service) RegisterNotifier(notifier notifiers.Notifier) error {
	_, exists := s.notifiers[notifier.Name()]
	if exists {
		return errors.New("redalert: notifier already existing on service. name: " + notifier.Name())
	}
	s.notifiers[notifier.Name()] = notifier
	return nil
}

func (s *Service) RegisterCheck(check *Check, sendAlerts []string, checkIdx int) error {
	err := check.AddNotifiers(s, sendAlerts)
	if err != nil {
		return err
	}
	s.checks[check.Data.ID] = check
	check.ConfigRank = checkIdx
	return nil
}
