package checks

import (
	"errors"
	"io/ioutil"
	"log"
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
	Address string
	log     *log.Logger
}

var WebPingerMetrics = map[string]MetricInfo{
	"latency": MetricInfo{
		Unit: "ms",
	},
}

func NewWebPinger(address string, logger *log.Logger) *WebPinger {
	return &WebPinger{address, logger}
}

var GlobalClient = http.Client{
	Timeout: time.Duration(10 * time.Second),
}

func (wp *WebPinger) Check() (Metrics, error) {

	metrics, err := wp.ping()
	if err != nil {
		// if the initial ping fails, retry after 5 seconds
		// the retry is to avoid noise from intermittent network/connection issues
		time.Sleep(5 * time.Second)
		return wp.ping()
	}

	return metrics, nil
}

func (wp *WebPinger) ping() (Metrics, error) {

	metrics := Metrics(make(map[string]float64))
	metrics["latency"] = float64(0)

	startTime := time.Now()
	wp.log.Println("GET", wp.Address)

	req, err := http.NewRequest("GET", wp.Address, nil)
	if err != nil {
		wp.log.Println("FAIL ", red, "OK", reset)
		return metrics, errors.New("web-ping: failed parsing url in http.NewRequest " + err.Error())
	}

	req.Header.Add("User-Agent", "Redalert/1.0")
	resp, err := GlobalClient.Do(req)
	if err != nil {
		wp.log.Println("FAIL ", red, "OK", reset)
		return metrics, errors.New("web-ping: failed client.Do " + err.Error())
	}

	_, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	endTime := time.Now()
	latency := endTime.Sub(startTime)
	metrics["latency"] = float64(latency.Seconds() * 1e3)

	wp.log.Println("Latency", white, metrics, reset)

	if err != nil {
		return metrics, errors.New("web-ping: failed reading body " + err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return metrics, errors.New("web-ping: non-200 status code. status code was " + strconv.Itoa(resp.StatusCode))
	}

	wp.log.Println("Status", green, "200 OK", reset)

	return metrics, nil
}

func (wp *WebPinger) MetricInfo(metric string) MetricInfo {
	return WebPingerMetrics[metric]
}

func (wp *WebPinger) RedAlertMessage() string {
	return "Uhoh, failed ping to " + wp.Address
}

func (wp *WebPinger) GreenAlertMessage() string {
	return "Woo-hoo, successful ping to" + wp.Address
}
