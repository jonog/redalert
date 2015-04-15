package core

import (
	"container/list"
	"log"
	"os"
	"os/signal"
	"sync"
)

var (
	green           = string([]byte{27, 91, 57, 55, 59, 52, 50, 109})
	red             = string([]byte{27, 91, 57, 55, 59, 52, 49, 109})
	reset           = string([]byte{27, 91, 48, 109})
	white           = string([]byte{27, 91, 57, 48, 59, 52, 55, 109})
	MaxEventsStored = 100
)

type Server struct {
	Name         string
	Address      string
	Interval     int
	Alerts       []Alert
	Log          *log.Logger
	service      *Service
	failCount    int
	LastEvent    *Event
	EventHistory *list.List
	wg           sync.WaitGroup
}

func NewServer(name, address string, interval int) *Server {
	var wg sync.WaitGroup
	return &Server{
		Name:         name,
		Address:      address,
		Interval:     interval,
		Alerts:       make([]Alert, 0),
		Log:          log.New(os.Stdout, name+" ", log.Ldate|log.Ltime),
		wg:           wg,
		EventHistory: list.New(),
	}
}

func (s *Server) AddAlerts(names []string) {
	for _, name := range names {
		s.Alerts = append(s.Alerts, s.service.getAlert(name))
	}
}

func (s *Server) Monitor() {

	s.service.wg.Add(1)
	s.wg.Add(1)

	stopScheduler := make(chan bool)
	s.SchedulePing(stopScheduler)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		for _ = range sigChan {
			stopScheduler <- true
			s.wg.Done()
		}
	}()

	s.wg.Wait()

	s.service.wg.Done()

}

func (s *Server) TriggerAlerts(event *Event) {

	go func() {

		var err error
		for _, alert := range s.Alerts {
			err = alert.Trigger(event)
			if err != nil {
				s.Log.Println(red, "CRITICAL: Failure triggering alert ["+alert.Name()+"]: ", err.Error())
			}
		}

	}()
}

func (s *Server) IncrFailCount() {
	s.failCount++
}
