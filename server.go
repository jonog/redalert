package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"
)

var (
	green = string([]byte{27, 91, 57, 55, 59, 52, 50, 109})
	red   = string([]byte{27, 91, 57, 55, 59, 52, 49, 109})
	reset = string([]byte{27, 91, 48, 109})
	white = string([]byte{27, 91, 57, 48, 59, 52, 55, 109})
)

type Server struct {
	name     string
	address  string
	interval int
	alerts   []Alert
	log      *log.Logger
	service  *Service
	wg       sync.WaitGroup
}

func (s *Service) AddServer(name string, address string, interval int, alertNames []string) {

	alerts := []Alert{}
	for _, alertName := range alertNames {
		alerts = append(alerts, s.GetAlert(alertName))
	}

	var wg sync.WaitGroup
	s.servers = append(s.servers, &Server{
		name:     name,
		address:  address,
		interval: interval,
		alerts:   alerts,
		log:      log.New(os.Stdout, name+" ", log.Ldate|log.Ltime),
		service:  s,
		wg:       wg,
	})

}

func (s *Server) Ping() error {

	s.log.Println("Pinging: ", s.name)
	resp, err := http.Get(s.address)
	if err != nil {
		return errors.New("redalert ping: failed http.Get " + err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("redalert ping: non-200 status code. status code was " + strconv.Itoa(resp.StatusCode))
	}
	s.log.Println(green, "OK", reset, s.name)

	return nil
}

func (s *Server) Monitor() {

	s.service.wg.Add(1)
	s.wg.Add(1)

	var err error
	ticker := time.NewTicker(time.Second * time.Duration(s.interval))
	go func() {

		var startTime time.Time
		var endTime time.Time

		for _ = range ticker.C {

			startTime = time.Now()
			err = s.Ping()
			endTime = time.Now()
			s.log.Println(white, "Analytics: ", endTime.Sub(startTime), reset)

			if err != nil {
				s.log.Println(red, "ERROR: ", err, reset, s.name)
				s.TriggerAlerts()
			}
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for _ = range c {
			s.wg.Done()
		}
	}()

	s.wg.Wait()

	s.service.wg.Done()

}

func (s *Server) TriggerAlerts() {

	event := &Event{server: s, time: time.Now()}

	var err error
	for _, alert := range s.alerts {
		err = alert.Trigger(event)
		if err != nil {
			s.log.Fatal(err)
		}
	}
}
