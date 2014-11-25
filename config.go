package main

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	Servers []ServerConfig `json:"servers"`
	Gmail   *GmailConfig   `json:"gmail,omitempty"`
	Slack   *SlackConfig   `json:"slack,omitempty"`
	Twilio  *TwilioConfig  `json:"twilio,omitempty"`
}

type ServerConfig struct {
	Name     string   `json:"name"`
	Address  string   `json:"address"`
	Interval int      `json:"interval"`
	Alerts   []string `json:"alerts"`
}

type GmailConfig struct {
	User                  string   `json:"user"`
	Pass                  string   `json:"pass"`
	NotificationAddresses []string `json:"notification_addresses"`
}

type SlackConfig struct {
	WebhookURL string `json:"webhook_url"`
}

type TwilioConfig struct {
	AccountSID          string   `json:"account_sid"`
	AuthToken           string   `json:"auth_token"`
	TwilioNumber        string   `json:"twilio_number"`
	NotificationNumbers []string `json:"notification_numbers"`
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
