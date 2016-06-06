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
	Command          string
	Shell            string
	OutputType       string
	ExpectedExitCode string
	ExpectedOutput   string
	log              *log.Logger
}

var CommandMetrics = map[string]MetricInfo{
	"execution_time": MetricInfo{
		Unit: "ms",
	},
}

type CommandConfig struct {
	Command          string `json:"command"`
	Shell            string `json:"shell"`
	OutputType       string `json:"output_type"`
	ExpectedExitCode string `json:"expected_exit_code"`
	ExpectedOutput   string `json:"expected_output"`
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
	if commandConfig.ExpectedExitCode == "" {
		commandConfig.ExpectedExitCode = "0"
	}
	return Checker(&Command{
		commandConfig.Command,
		commandConfig.Shell,
		commandConfig.OutputType,
		commandConfig.ExpectedExitCode,
		commandConfig.ExpectedOutput,
		logger}), nil
}

func (c *Command) Check() (Metrics, error) {

	metrics := Metrics(make(map[string]*float64))
	executionTime := float64(0)

	c.log.Println("Run command:", c.Command, "using shell:", c.Shell)

	startTime := time.Now()

	cmd := exec.Command(c.Shell, "-c", c.Command)
	out, err := cmd.CombinedOutput()
	endTime := time.Now()

	exitCode := "0"
	if err != nil {
		exitCode = strings.Replace(err.Error(), "exit status ", "", -1)
	}
	c.log.Println("Command finished with exit code:", exitCode)
	outStr := bytes.NewBuffer(out).String()

	executionTimeCalc := endTime.Sub(startTime)
	executionTime = float64(executionTimeCalc.Seconds() * 1e3)
	c.log.Println("Execution Time", utils.White, executionTime, utils.Reset)
	metrics["execution_time"] = &executionTime

	if c.ExpectedExitCode != "" && c.ExpectedExitCode != exitCode {
		return metrics, errors.New("command: unexcpected exit code:" + exitCode)
	}

	if c.OutputType == "number" {
		f, err := strconv.ParseFloat(strings.TrimSpace(outStr), 64)
		if err != nil {
			return metrics, errors.New("command: error while parsing number: " + err.Error())
		}
		metrics["output"] = &f
	}

	if c.ExpectedOutput != "" && c.ExpectedOutput != outStr {
		return metrics, errors.New("command: unexcpected output:" + outStr)
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
