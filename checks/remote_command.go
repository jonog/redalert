package checks

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/jonog/redalert/data"
	"github.com/jonog/redalert/utils"
)

func init() {
	Register("remote-command", NewRemoteCommand)
}

type RemoteCommand struct {
	Command        string
	OutputType     string
	SSHAuthOptions SSHAuthOptions
	log            *log.Logger
}

var RemoteCommandMetrics = map[string]MetricInfo{
	"execution_time": {
		Unit: "ms",
	},
}

type RemoteCommandConfig struct {
	Command        string         `json:"command"`
	OutputType     string         `json:"output_type"`
	SSHAuthOptions SSHAuthOptions `json:"ssh_auth_options"`
}

var NewRemoteCommand = func(config Config, logger *log.Logger) (Checker, error) {
	var commandConfig RemoteCommandConfig
	err := json.Unmarshal([]byte(config.Config), &commandConfig)
	if err != nil {
		return nil, err
	}
	if commandConfig.Command == "" {
		return nil, errors.New("remote-command: command to run cannot be blank")
	}
	if commandConfig.OutputType != "" && commandConfig.OutputType != "number" {
		return nil, errors.New("remote-command: invalid output type")
	}
	if commandConfig.OutputType == "number" {
		RemoteCommandMetrics["output"] = MetricInfo{
			Unit: "unit",
		}
	}
	return Checker(&RemoteCommand{
		commandConfig.Command,
		commandConfig.OutputType,
		commandConfig.SSHAuthOptions,
		logger}), nil
}

func (c *RemoteCommand) Check() (data.CheckResponse, error) {

	response := data.CheckResponse{
		Metrics:  data.Metrics(make(map[string]*float64)),
		Metadata: make(map[string]string),
	}
	executionTime := float64(0)

	c.log.Println("Connect via SSH to...")
	auth := NewSSHAuthenticator(c.log, SSHAuthOptions{Password: c.SSHAuthOptions.Password, Key: c.SSHAuthOptions.Key})
	if len(auth.auths) == 0 {
		return response, errors.New("remote-command: no SSH authentication methods available")
	}
	defer auth.Cleanup()

	var sshPortStr string
	if c.SSHAuthOptions.Port == 0 {
		sshPortStr = "22"
	} else {
		sshPortStr = strconv.Itoa(c.SSHAuthOptions.Port)
	}
	client, err := ssh.Dial("tcp", c.SSHAuthOptions.Host+":"+sshPortStr, &ssh.ClientConfig{
		User: c.SSHAuthOptions.User,
		Auth: auth.auths,
	})
	if err != nil {
		return response, fmt.Errorf("remote-command: error dialing ssh. err: %v", err)
	}
	defer client.Close()

	startTime := time.Now()
	out, err := runCommand(client, c.Command)
	response.Response = out
	endTime := time.Now()

	executionTimeCalc := endTime.Sub(startTime)
	executionTime = float64(executionTimeCalc.Seconds() * 1e3)
	c.log.Println("Execution Time", utils.White, executionTime, utils.Reset)
	response.Metrics["execution_time"] = &executionTime

	if err != nil {
		if exitErr, ok := err.(*ssh.ExitError); ok {
			response.Metadata["exit_status"] = strconv.Itoa(exitErr.ExitStatus())
			return response, nil
		}
		return response, errors.New("remote-command: " + err.Error())
	}

	response.Metadata["exit_status"] = "0"

	if c.OutputType == "number" {
		numberStr := bytes.NewBuffer(out).String()
		f, err := strconv.ParseFloat(strings.TrimSpace(numberStr), 64)
		if err != nil {
			return response, errors.New("remote-command: error while parsing number: " + err.Error())
		}
		response.Metrics["output"] = &f
	}

	c.log.Println("Output: ", fmt.Sprintf("%s", out))
	return response, nil

}

func (c *RemoteCommand) MetricInfo(metric string) MetricInfo {
	return RemoteCommandMetrics[metric]
}

func (c *RemoteCommand) MessageContext() string {
	return c.Command
}
