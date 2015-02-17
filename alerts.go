package main

import (
	"log"
	"os"
)

type Alert interface {
	Trigger(*Event) error
}

func (s *Service) ConfigureAlerts() {

	logger := log.New(os.Stdout, "Setup ", log.Ldate|log.Ltime)

	s.alerts = make(map[string]Alert)

	s.alerts["stderr"] = StandardError{}

	if s.config.Slack == nil || s.config.Slack.WebhookURL == "" {
		logger.Println("Slack is not configured")
	} else {
		s.alerts["slack"] = SlackWebhook{
			url:       s.config.Slack.WebhookURL,
			channel:   s.config.Slack.Channel,
			username:  s.config.Slack.Username,
			iconEmoji: s.config.Slack.IconEmoji,
		}
	}

	if s.config.Gmail == nil || s.config.Gmail.User == "" || s.config.Gmail.Pass == "" || len(s.config.Gmail.NotificationAddresses) == 0 {
		logger.Println("Gmail is not configured")
	} else {
		s.alerts["gmail"] = Gmail{
			user: s.config.Gmail.User,
			pass: s.config.Gmail.Pass,
			notificationAddresses: s.config.Gmail.NotificationAddresses,
		}
	}

	if s.config.Twilio == nil || s.config.Twilio.AccountSID == "" || s.config.Twilio.AuthToken == "" || len(s.config.Twilio.NotificationNumbers) == 0 || s.config.Twilio.TwilioNumber == "" {
		logger.Println("Twilio is not configured")
	} else {
		s.alerts["twilio"] = Twilio{
			accountSid:   s.config.Twilio.AccountSID,
			authToken:    s.config.Twilio.AuthToken,
			phoneNumbers: s.config.Twilio.NotificationNumbers,
			twilioNumber: s.config.Twilio.TwilioNumber,
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
