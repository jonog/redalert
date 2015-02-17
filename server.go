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
	Name      string
	Address   string
	Interval  int
	alerts    []Alert
	log       *log.Logger
	service   *Service
	failCount int
	LastEvent *Event
	wg        sync.WaitGroup
}

func (s *Service) AddServer(name string, address string, interval int, alertNames []string) {

	alerts := []Alert{}
	for _, alertName := range alertNames {
		alerts = append(alerts, s.GetAlert(alertName))
	}

	var wg sync.WaitGroup
	s.servers = append(s.servers, &Server{
		Name:     name,
		Address:  address,
		Interval: interval,
		alerts:   alerts,
		log:      log.New(os.Stdout, name+" ", log.Ldate|log.Ltime),
		service:  s,
		wg:       wg,
	})

}

func (s *Server) Ping() (time.Duration, error) {

	startTime := time.Now()
	s.log.Println("Pinging: ", s.Name)

	req, err := http.NewRequest("GET", s.Address, nil)
	if err != nil {
		return 0, errors.New("redalert ping: failed parsing url in http.NewRequest " + err.Error())
	}

	req.Header.Add("User-Agent", "Redalert/1.0")
	client := &http.Client{}
	resp, err := client.Do(req)

	endTime := time.Now()
	latency := endTime.Sub(startTime)
	s.log.Println(white, "Analytics: ", latency, reset)

	if err != nil {
		return latency, errors.New("redalert ping: failed client.Do " + err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return latency, errors.New("redalert ping: non-200 status code. status code was " + strconv.Itoa(resp.StatusCode))
	}
	s.log.Println(green, "OK", reset, s.Name)

	return latency, nil
}

func (s *Server) SchedulePing(stopChan chan bool) {

	go func() {

		var err error
		var event *Event
		var latency time.Duration

		originalDelay := time.Second * time.Duration(s.Interval)
		delay := time.Second * time.Duration(s.Interval)

		for {

			latency, err = s.Ping()

			if err != nil {

				s.log.Println(red, "ERROR: ", err, reset, s.Name)
				event = NewRedAlert(s, latency)
				s.LastEvent = event

				s.TriggerAlerts(event)

				s.IncrFailCount()
				if s.failCount > 0 {
					delay = time.Second * time.Duration(s.failCount*s.Interval)
				}

			} else {

				event = NewGreenAlert(s, latency)
				isRedalertRecovery := s.LastEvent != nil && s.LastEvent.isRedAlert()
				s.LastEvent = event
				if isRedalertRecovery {
					s.log.Println(green, "RECOVERY: ", reset, s.Name)
					s.TriggerAlerts(event)
				}

				delay = originalDelay
				s.failCount = 0

			}

			select {
			case <-time.After(delay):
			case <-stopChan:
				return
			}
		}
	}()

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

	// TODO: shift to a queue
	var err error
	for _, alert := range s.alerts {
		err = alert.Trigger(event)
		if err != nil {
			s.log.Fatal(err)
		}
	}
}

func (s *Server) IncrFailCount() {
	s.failCount++
}
