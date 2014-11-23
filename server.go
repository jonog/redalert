package main

import (
	"errors"
	"log"
	"net/http"
	"os"
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
}

func (s *Service) AddServer(name string, address string, interval int, alertNames []string) {

	alerts := []Alert{}
	for _, alertName := range alertNames {
		alerts = append(alerts, s.GetAlert(alertName))
	}

	s.servers = append(s.servers, &Server{
		name:     name,
		address:  address,
		interval: interval,
		alerts:   alerts,
		log:      log.New(os.Stdout, name+" ", log.Ldate|log.Ltime),
	})

}

func (s *Server) Ping() error {

	s.log.Println("Pinging: ", s.name)
	resp, err := http.Get(s.address)
	if err != nil {
		return errors.New("Error retrieving from URL")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("Invalid status code")
	}
	s.log.Println(green, "OK", reset, s.name)

	return nil
}

func (s *Server) Monitor() {

	var err error
	ticker := time.NewTicker(time.Second * time.Duration(s.interval))
	go func() {
		for _ = range ticker.C {
			err = s.Ping()
			if err != nil {
				s.log.Println(red, "ERROR", reset, s.name)
				s.TriggerAlerts()
			}
		}
	}()

	block := make(chan bool)
	<-block
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
