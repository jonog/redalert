package checks

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/jonog/redalert/data"
	"github.com/jonog/redalert/utils"
)

func init() {
	Register("web-ping", NewWebPinger)
}

type WebPinger struct {
	Address string
	Headers map[string]string
	log     *log.Logger
}

var WebPingerMetrics = map[string]MetricInfo{
	"latency": MetricInfo{
		Unit: "ms",
	},
}

type WebPingerConfig struct {
	Address string            `json:"address"`
	Headers map[string]string `json:"headers"`
}

var NewWebPinger = func(config Config, logger *log.Logger) (Checker, error) {
	var webPingerConfig WebPingerConfig
	err := json.Unmarshal([]byte(config.Config), &webPingerConfig)
	if err != nil {
		return nil, err
	}
	if webPingerConfig.Address == "" {
		return nil, errors.New("web-ping: address to ping cannot be blank")
	}
	return Checker(&WebPinger{webPingerConfig.Address, webPingerConfig.Headers, logger}), nil
}

var GlobalClient = http.Client{
	Timeout: time.Duration(10 * time.Second),
}

func (wp *WebPinger) Check() (data.CheckResponse, error) {
	metadata := make(map[string]string)
	metrics, b, statusCode, err := wp.ping()
	if err != nil {
		// if the initial ping fails, retry after 5 seconds
		// the retry is to avoid noise from intermittent network/connection issues
		time.Sleep(5 * time.Second)
		var secondMetrics map[string]*float64
		var secondStatusCode int
		var secondB []byte
		secondMetrics, secondB, secondStatusCode, err = wp.ping()
		metadata["status_code"] = strconv.Itoa(secondStatusCode)
		return data.CheckResponse{Metrics: secondMetrics, Metadata: metadata, Response: secondB}, err
	}
	metadata["status_code"] = strconv.Itoa(statusCode)
	return data.CheckResponse{Metrics: metrics, Metadata: metadata, Response: b}, nil
}

func (wp *WebPinger) ping() (data.Metrics, []byte, int, error) {

	metrics := data.Metrics(make(map[string]*float64))
	var b []byte

	latency := float64(0)
	defer func() {
		metrics["latency"] = &latency
	}()

	startTime := time.Now()
	wp.log.Println("GET", wp.Address)

	req, err := http.NewRequest("GET", wp.Address, nil)
	if err != nil {
		return metrics, b, 0, errors.New("web-ping: failed parsing url in http.NewRequest " + err.Error())
	}

	req.Header.Add("User-Agent", "Redalert/1.0")
	for k, v := range wp.Headers {
		req.Header.Add(k, v)
	}

	resp, err := GlobalClient.Do(req)
	if err != nil {
		return metrics, b, 0, errors.New("web-ping: failed client.Do " + err.Error())
	}

	b, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	endTime := time.Now()
	latencyCalc := endTime.Sub(startTime)
	latency = float64(latencyCalc.Seconds() * 1e3)

	wp.log.Println("Latency", utils.White, latency, utils.Reset)

	if err != nil {
		return metrics, b, resp.StatusCode, errors.New("web-ping: failed reading body " + err.Error())
	}

	return metrics, b, resp.StatusCode, nil
}

func (wp *WebPinger) MetricInfo(metric string) MetricInfo {
	return WebPingerMetrics[metric]
}

func (wp *WebPinger) MessageContext() string {
	return wp.Address
}
