package main

import (
	"os"
	"os/signal"
	"sync"
)

type Service struct {
	servers []*Server
	alerts  map[string]Alert
	wg      sync.WaitGroup
}

func (s *Service) Start() {

	// use this to keep the service running, even if no monitoring is occuring
	s.wg.Add(1)

	for _, server := range s.servers {
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

func main() {

	service := new(Service)

	config, err := ReadConfigFile()
	if err != nil {
		panic("Missing or invalid config")
	}

	service.SetupAlerts(config)

	for _, sc := range config.Servers {
		service.AddServer(sc.Name, sc.Address, sc.Interval, sc.Alerts)
	}

	service.Start()
	service.wg.Wait()

}
