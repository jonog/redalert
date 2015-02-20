package main

import (
	"log"
	"os"
)

type Alert interface {
	Trigger(*Event) error
	Name() string
}

func (s *Service) ConfigureAlerts() {

	logger := log.New(os.Stdout, "Setup ", log.Ldate|log.Ltime)

	s.alerts = make(map[string]Alert)

	s.alerts["stderr"] = NewStandardError()

	if s.config.Slack == nil || s.config.Slack.WebhookURL == "" {
		logger.Println("Slack is not configured")
	} else {
		s.alerts["slack"] = NewSlackWebhook(s.config.Slack)

	}

	if s.config.Gmail == nil || s.config.Gmail.User == "" || s.config.Gmail.Pass == "" || len(s.config.Gmail.NotificationAddresses) == 0 {
		logger.Println("Gmail is not configured")
	} else {
		s.alerts["gmail"] = NewGmail(s.config.Gmail)
	}

	if s.config.Twilio == nil || s.config.Twilio.AccountSID == "" || s.config.Twilio.AuthToken == "" || len(s.config.Twilio.NotificationNumbers) == 0 || s.config.Twilio.TwilioNumber == "" {
		logger.Println("Twilio is not configured")
	} else {
		s.alerts["twilio"] = NewTwilio(s.config.Twilio)
	}

}

func (s *Service) GetAlert(name string) Alert {
	alert, ok := s.alerts[name]
	if !ok {
		panic("Alert has not been registered!")
	}
	return alert
}
