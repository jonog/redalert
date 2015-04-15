package core

import (
	"os"
	"os/signal"
	"sync"
)

type Service struct {
	Servers []*Server
	Alerts  map[string]Alert
	wg      sync.WaitGroup
}

func NewService() *Service {
	return &Service{Alerts: make(map[string]Alert)}
}

func (s *Service) Start() {

	// use this to keep the service running, even if no monitoring is occuring
	s.wg.Add(1)

	for _, server := range s.Servers {
		go server.Monitor()
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for _ = range c {
			s.wg.Done()
		}
	}()

}

func (s *Service) KeepRunning() {
	s.wg.Wait()
}

func (s *Service) AddServer(name string, address string, interval int, alertNames []string) {
	server := NewServer(name, address, interval)
	server.service = s
	server.AddAlerts(alertNames)
	s.Servers = append(s.Servers, server)
}

func (s *Service) getAlert(name string) Alert {
	alert, ok := s.Alerts[name]
	if !ok {
		panic("Alert has not been registered!")
	}
	return alert
}
