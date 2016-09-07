package checks

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/jonog/redalert/data"
	"github.com/jonog/redalert/utils"
)

func init() {
	Register("test-report", NewTestReport)
}

type TestReport struct {
	Command string
	Shell   string
	log     *log.Logger
}

var TestReportMetrics = map[string]MetricInfo{
	"execution_time": {
		Unit: "ms",
	},
}

type TestReportConfig struct {
	Command string `json:"command"`
	Shell   string `json:"shell"`
}

var NewTestReport = func(config Config, logger *log.Logger) (Checker, error) {
	var testReportConfig TestReportConfig
	err := json.Unmarshal([]byte(config.Config), &testReportConfig)
	if err != nil {
		return nil, err
	}
	if testReportConfig.Command == "" {
		return nil, errors.New("command: command to run cannot be blank")
	}
	if testReportConfig.Shell == "" {
		testReportConfig.Shell = "sh"
	}
	return Checker(&TestReport{
		testReportConfig.Command,
		testReportConfig.Shell,
		logger}), nil
}

func (c *TestReport) Check() (data.CheckResponse, error) {

	response := data.CheckResponse{
		Metrics:  data.Metrics(make(map[string]*float64)),
		Metadata: make(map[string]string),
	}
	executionTime := float64(0)

	c.log.Println("Run test-suite via:", c.Command, "using shell:", c.Shell)

	startTime := time.Now()
	// ignore error here and rely on xml parsing error
	out, err := exec.Command(c.Shell, "-c", c.Command).Output()
	fmt.Println(err)
	response.Response = out
	endTime := time.Now()

	executionTimeCalc := endTime.Sub(startTime)
	executionTime = float64(executionTimeCalc.Seconds() * 1e3)
	c.log.Println("Execution Time", utils.White, executionTime, utils.Reset)
	response.Metrics["execution_time"] = &executionTime

	var testReport Testsuite
	xmlErr := xml.Unmarshal(out, &testReport)
	if xmlErr != nil {
		return response, errors.New("test-suite: invalid junit xml: " + xmlErr.Error())
	}
	testCount := float64(testReport.Tests)
	response.Metrics["test_count"] = &testCount
	failureCount := float64(testReport.Failures)
	if failureCount > 0 {
		response.Metadata["status"] = "FAILING"
	} else {
		response.Metadata["status"] = "PASSING"
	}
	response.Metrics["failure_count"] = &failureCount

	skippedCountInt := 0
	for _, test := range testReport.Testcases {
		testCase := *test
		if testCase.Skipped != nil {
			skippedCountInt++
		}
	}
	skippedCount := float64(skippedCountInt)
	response.Metrics["skipped_count"] = &skippedCount

	passCount := float64(testCount - failureCount - skippedCount)
	response.Metrics["pass_count"] = &passCount

	c.log.Println("Report: ", fmt.Sprintf("%s", out))

	return response, nil
}

func (c *TestReport) MetricInfo(metric string) MetricInfo {
	return TestReportMetrics[metric]
}

func (c *TestReport) MessageContext() string {
	return c.Command
}

type Testsuite struct {
	Name      string      `xml:"name,attr"`
	Tests     int         `xml:"tests,attr"`
	Failures  int         `xml:"failures,attr"`
	Errors    int         `xml:"errors,attr"`
	Timestamp string      `xml:"timestamp,attr"`
	Time      float64     `xml:"time,attr"`
	Hostname  string      `xml:"hostname,attr"`
	Testcases []*TestCase `xml:"testcase"`
}

type TestCase struct {
	Name      string    `xml:"name,attr"`
	Time      float64   `xml:"time,attr"`
	Classname string    `xml:"classname,attr"`
	Failure   *Failure  `xml:"failure"`
	Skipped   *struct{} `xml:"skipped"`
}

type Failure struct {
	Type    string `xml:"type,attr"`
	Message string `xml:"message,attr"`
}
