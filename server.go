package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"time"
)

type Server struct {
	name     string
	address  string
	interval int
	actions  []Action
	log      *log.Logger
}

func (s *Service) AddServer(name string, address string, interval int, actionNames []string) {

	actions := []Action{}
	for _, actionName := range actionNames {
		actions = append(actions, s.GetAction(actionName))
	}

	s.servers = append(s.servers, &Server{
		name:     name,
		address:  address,
		interval: interval,
		actions:  actions,
		log:      log.New(os.Stdout, name+" ", log.Ldate|log.Ltime|log.Lshortfile),
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
	s.log.Println("OK", s.name)

	return nil
}

func (s *Server) Monitor() {

	var err error
	ticker := time.NewTicker(time.Second * time.Duration(s.interval))
	go func() {
		for _ = range ticker.C {
			err = s.Ping()
			if err != nil {
				s.log.Println("ERROR", s.name)
				s.TriggerActions()
			}
		}
	}()

	block := make(chan bool)
	<-block
}

func (s *Server) TriggerActions() {

	var err error
	for _, alert := range s.actions {
		err = alert.Send(s)
		if err != nil {
			s.log.Fatal(err)
		}
	}
}
