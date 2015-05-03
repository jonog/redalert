package checks

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

var (
	green = string([]byte{27, 91, 57, 55, 59, 52, 50, 109})
	red   = string([]byte{27, 91, 57, 55, 59, 52, 49, 109})
	reset = string([]byte{27, 91, 48, 109})
	white = string([]byte{27, 91, 57, 48, 59, 52, 55, 109})
)

type WebPinger struct {
	Identifier string
	Address    string
}

func NewWebPinger(identifier, address string) *WebPinger {
	return &WebPinger{identifier, address}
}

var GlobalClient = http.Client{
	Timeout: time.Duration(10 * time.Second),
}

func (wp *WebPinger) Check() (map[string]float64, error) {

	metrics := make(map[string]float64)
	metrics["latency"] = float64(0)

	startTime := time.Now()
	fmt.Println(wp.Identifier, " : Pinging")

	req, err := http.NewRequest("GET", wp.Address, nil)
	if err != nil {
		fmt.Println(wp.Identifier, " : FAIL ", red, "OK", reset)
		return metrics, errors.New("redalert ping: failed parsing url in http.NewRequest " + err.Error())
	}

	req.Header.Add("User-Agent", "Redalert/1.0")
	resp, err := GlobalClient.Do(req)
	if err != nil {
		fmt.Println(wp.Identifier, " : FAIL ", red, "OK", reset)
		return metrics, errors.New("redalert ping: failed client.Do " + err.Error())
	}

	_, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	endTime := time.Now()
	latency := endTime.Sub(startTime)
	metrics["latency"] = float64(latency.Seconds() * 1e3)

	fmt.Println(wp.Identifier, " : Analytics ", white, metrics, reset)

	if err != nil {
		return metrics, errors.New("redalert ping: failed reading body " + err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return metrics, errors.New("redalert ping: non-200 status code. status code was " + strconv.Itoa(resp.StatusCode))
	}

	fmt.Println(wp.Identifier, " : Analytics ", green, "OK", reset)

	return metrics, nil
}

func (wp *WebPinger) RedAlertMessage() string {
	return "Uhoh, failed ping to" + wp.Address
}

func (wp *WebPinger) GreenAlertMessage() string {
	return "Woo-hoo, successful ping to" + wp.Address
}
