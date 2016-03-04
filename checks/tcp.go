package checks

import (
	"encoding/json"
	"errors"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/jonog/redalert/utils"
)

func init() {
	Register("tcp", NewTCP)
}

type TCP struct {
	Host string
	Port int
	log  *log.Logger
}

var TCPMetrics = map[string]MetricInfo{
	"latency": MetricInfo{
		Unit: "ms",
	},
}

type TCPConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

var NewTCP = func(config Config, logger *log.Logger) (Checker, error) {
	var tcpConfig TCPConfig
	err := json.Unmarshal([]byte(config.Config), &tcpConfig)
	if err != nil {
		return nil, err
	}
	if tcpConfig.Host == "" {
		return nil, errors.New("tcp: host to connect to cannot be blank")
	}
	if tcpConfig.Port == 0 {
		return nil, errors.New("tcp: port to connect to cannot be zero")
	}
	return Checker(&TCP{tcpConfig.Host, tcpConfig.Port, logger}), nil
}

func (t *TCP) Check() (Metrics, error) {

	metrics := Metrics(make(map[string]*float64))
	latency := float64(0)

	startTime := time.Now()

	t.log.Println("Establish TCP with", address(t.Host, t.Port))
	conn, err := net.Dial("tcp", address(t.Host, t.Port))

	endTime := time.Now()
	latencyCalc := endTime.Sub(startTime)
	latency = float64(latencyCalc.Seconds() * 1e3)
	t.log.Println("Latency", utils.White, latency, utils.Reset)
	metrics["latency"] = &latency

	if err != nil {
		return metrics, errors.New("tcp: " + formatError(err))
	}

	conn.Close()
	return metrics, nil
}

func (t *TCP) MetricInfo(metric string) MetricInfo {
	return TCPMetrics[metric]
}

func (t *TCP) MessageContext() string {
	return address(t.Host, t.Port)
}

func formatError(err error) string {
	if strings.Contains(err.Error(), "connection refused") {
		return "connection error"
	}
	return err.Error()
}

func address(host string, port int) string {
	return host + ":" + strconv.Itoa(port)
}
