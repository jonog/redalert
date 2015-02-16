package main

import (
	"log"
	"os"
)

type Alert interface {
	Trigger(*Event) error
}

func (s *Service) SetupAlerts(config *Config) {

	logger := log.New(os.Stdout, "Setup ", log.Ldate|log.Ltime)

	s.alerts = make(map[string]Alert)

	s.alerts["stderr"] = StandardError{}

	if config.Slack == nil || config.Slack.WebhookURL == "" {
		logger.Println("Slack is not configured")
	} else {
		s.alerts["slack"] = SlackWebhook{
			url:       config.Slack.WebhookURL,
			channel:   config.Slack.Channel,
			username:  config.Slack.Username,
			iconEmoji: config.Slack.IconEmoji,
		}
	}

	if config.Gmail == nil || config.Gmail.User == "" || config.Gmail.Pass == "" || len(config.Gmail.NotificationAddresses) == 0 {
		logger.Println("Gmail is not configured")
	} else {
		s.alerts["gmail"] = Gmail{
			user: config.Gmail.User,
			pass: config.Gmail.Pass,
			notificationAddresses: config.Gmail.NotificationAddresses,
		}
	}

	if config.Twilio == nil || config.Twilio.AccountSID == "" || config.Twilio.AuthToken == "" || len(config.Twilio.NotificationNumbers) == 0 || config.Twilio.TwilioNumber == "" {
		logger.Println("Twilio is not configured")
	} else {
		s.alerts["twilio"] = Twilio{
			accountSid:   config.Twilio.AccountSID,
			authToken:    config.Twilio.AuthToken,
			phoneNumbers: config.Twilio.NotificationNumbers,
			twilioNumber: config.Twilio.TwilioNumber,
		}
	}

}

func (s *Service) GetAlert(name string) Alert {
	alert, ok := s.alerts[name]
	if !ok {
		panic("Alert has not been registered!")
	}
	return alert
}
