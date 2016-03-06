package checks

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/jonog/redalert/utils"
)

func init() {
	Register("command", NewCommand)
}

type Command struct {
	Command    string
	Shell      string
	OutputType string
	log        *log.Logger
}

var CommandMetrics = map[string]MetricInfo{
	"execution_time": MetricInfo{
		Unit: "ms",
	},
}

type CommandConfig struct {
	Command    string `json:"command"`
	Shell      string `json:"shell"`
	OutputType string `json:"output_type"`
}

var NewCommand = func(config Config, logger *log.Logger) (Checker, error) {
	var commandConfig CommandConfig
	err := json.Unmarshal([]byte(config.Config), &commandConfig)
	if err != nil {
		return nil, err
	}
	if commandConfig.Command == "" {
		return nil, errors.New("command: command to run cannot be blank")
	}
	if commandConfig.Shell == "" {
		commandConfig.Shell = "sh"
	}
	if commandConfig.OutputType != "" && commandConfig.OutputType != "number" {
		return nil, errors.New("command: invalid output type")
	}
	if commandConfig.OutputType == "number" {
		CommandMetrics["output"] = MetricInfo{
			Unit: "unit",
		}
	}
	return Checker(&Command{
		commandConfig.Command,
		commandConfig.Shell,
		commandConfig.OutputType,
		logger}), nil
}

func (c *Command) Check() (Metrics, error) {

	metrics := Metrics(make(map[string]*float64))
	executionTime := float64(0)

	c.log.Println("Run command:", c.Command, "using shell:", c.Shell)

	startTime := time.Now()
	out, err := exec.Command(c.Shell, "-c", c.Command).Output()
	endTime := time.Now()

	executionTimeCalc := endTime.Sub(startTime)
	executionTime = float64(executionTimeCalc.Seconds() * 1e3)
	c.log.Println("Execution Time", utils.White, executionTime, utils.Reset)
	metrics["execution_time"] = &executionTime

	if err != nil {
		return metrics, errors.New("command: " + err.Error())
	}

	if c.OutputType == "number" {
		numberStr := bytes.NewBuffer(out).String()
		f, err := strconv.ParseFloat(strings.TrimSpace(numberStr), 64)
		if err != nil {
			return metrics, errors.New("command: error while parsing number: " + err.Error())
		}
		metrics["output"] = &f
	}

	c.log.Println("Output: ", fmt.Sprintf("%s", out))
	return metrics, nil

}

func (c *Command) MetricInfo(metric string) MetricInfo {
	return CommandMetrics[metric]
}

func (c *Command) MessageContext() string {
	return c.Command
}
