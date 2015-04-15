package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/jonog/redalert/alerts"
	"github.com/jonog/redalert/core"
)

type Config struct {
	Servers []ServerConfig       `json:"servers"`
	Gmail   *alerts.GmailConfig  `json:"gmail,omitempty"`
	Slack   *alerts.SlackConfig  `json:"slack,omitempty"`
	Twilio  *alerts.TwilioConfig `json:"twilio,omitempty"`
}

type ServerConfig struct {
	Name     string   `json:"name"`
	Address  string   `json:"address"`
	Interval int      `json:"interval"`
	Alerts   []string `json:"alerts"`
}

func ReadConfigFile() (*Config, error) {
	file, err := ioutil.ReadFile("config.json")
	if err != nil {
		return nil, err
	}
	var config Config
	err = json.Unmarshal(file, &config)
	return &config, err
}

func ConfigureStdErr(s *core.Service) {
	s.Alerts["stderr"] = alerts.NewStandardError()
}

func ConfigureGmail(s *core.Service, config *alerts.GmailConfig) {
	if config == nil || config.User == "" || config.Pass == "" || len(config.NotificationAddresses) == 0 {
		fmt.Println("Gmail is not configured")
	} else {
		s.Alerts["gmail"] = alerts.NewGmail(config)
	}
}

func ConfigureSlack(s *core.Service, config *alerts.SlackConfig) {
	if config == nil || config.WebhookURL == "" {
		fmt.Println("Slack is not configured")
	} else {
		s.Alerts["slack"] = alerts.NewSlackWebhook(config)
	}
}

func ConfigureTwilio(s *core.Service, config *alerts.TwilioConfig) {
	if config == nil || config.AccountSID == "" || config.AuthToken == "" || len(config.NotificationNumbers) == 0 || config.TwilioNumber == "" {
		fmt.Println("Twilio is not configured")
	} else {
		s.Alerts["twilio"] = alerts.NewTwilio(config)
	}
}
