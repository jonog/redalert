package core

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

var GlobalClient = http.Client{
	Timeout: time.Duration(10 * time.Second),
}

func (s *Server) Ping() (time.Duration, error) {

	startTime := time.Now()
	s.Log.Println("Pinging: ", s.Name)

	req, err := http.NewRequest("GET", s.Address, nil)
	if err != nil {
		return 0, errors.New("redalert ping: failed parsing url in http.NewRequest " + err.Error())
	}

	req.Header.Add("User-Agent", "Redalert/1.0")
	resp, err := GlobalClient.Do(req)
	if err != nil {
		return 0, errors.New("redalert ping: failed client.Do " + err.Error())
	}

	_, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	endTime := time.Now()
	latency := endTime.Sub(startTime)
	s.Log.Println(white, "Analytics: ", latency, reset)

	if err != nil {
		return latency, errors.New("redalert ping: failed reading body " + err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return latency, errors.New("redalert ping: non-200 status code. status code was " + strconv.Itoa(resp.StatusCode))
	}

	s.Log.Println(green, "OK", reset, s.Name)
	return latency, nil
}
