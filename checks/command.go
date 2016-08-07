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
	"syscall"
	"time"

	"github.com/jonog/redalert/data"
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
	"execution_time": {
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

func (c *Command) Check() (data.CheckResponse, error) {

	response := data.CheckResponse{
		Metrics:  data.Metrics(make(map[string]*float64)),
		Metadata: make(map[string]string),
	}
	executionTime := float64(0)

	c.log.Println("Run command:", c.Command, "using shell:", c.Shell)

	startTime := time.Now()
	out, err := exec.Command(c.Shell, "-c", c.Command).Output()
	response.Response = out
	endTime := time.Now()

	executionTimeCalc := endTime.Sub(startTime)
	executionTime = float64(executionTimeCalc.Seconds() * 1e3)
	c.log.Println("Execution Time", utils.White, executionTime, utils.Reset)
	response.Metrics["execution_time"] = &executionTime

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			waitStatus := exitError.Sys().(syscall.WaitStatus)
			response.Metadata["exit_status"] = strconv.Itoa(waitStatus.ExitStatus())
			return response, nil
		}
		return response, errors.New("command: " + err.Error())
	}

	response.Metadata["exit_status"] = "0"

	if c.OutputType == "number" {
		numberStr := bytes.NewBuffer(out).String()
		f, err := strconv.ParseFloat(strings.TrimSpace(numberStr), 64)
		if err != nil {
			return response, errors.New("command: error while parsing number: " + err.Error())
		}
		response.Metrics["output"] = &f
	}

	c.log.Println("Output: ", fmt.Sprintf("%s", out))
	return response, nil

}

func (c *Command) MetricInfo(metric string) MetricInfo {
	return CommandMetrics[metric]
}

func (c *Command) MessageContext() string {
	return c.Command
}
