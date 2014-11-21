package main

import (
	"log"
	"os"
)

type Action interface {
	Send(*Server) error
}

func (s *Service) SetupActions() {

	logger := log.New(os.Stdout, "Setup ", log.Ldate|log.Ltime)

	s.actions = make(map[string]Action)

	s.actions["stderr"] = StandardError{}

	if os.Getenv("RA_SLACK_URL") == "" {
		logger.Println("Slack is not configured")
	} else {
		s.actions["slack"] = SlackWebhook{url: os.Getenv("RA_SLACK_URL")}
	}

	if os.Getenv("RA_GMAIL_USER") == "" || os.Getenv("RA_GMAIL_PASS") == "" || os.Getenv("RA_GMAIL_NOTIFICATION_ADDRESS") == "" {
		logger.Println("Email is not configured")
	} else {
		s.actions["email"] = Email{
			user:                os.Getenv("RA_GMAIL_USER"),
			pass:                os.Getenv("RA_GMAIL_PASS"),
			notificationAddress: os.Getenv("RA_GMAIL_NOTIFICATION_ADDRESS"),
		}
	}

	if os.Getenv("RA_TWILIO_ACCOUNT_SID") == "" || os.Getenv("RA_TWILIO_AUTH_TOKEN") == "" || os.Getenv("RA_TWILIO_PHONE_NUMBER") == "" || os.Getenv("RA_TWILIO_TWILIO_NUMBER") == "" {
		logger.Println("SMS is not configured")
	} else {
		s.actions["sms"] = SMS{
			accountSid:   os.Getenv("RA_TWILIO_ACCOUNT_SID"),
			authToken:    os.Getenv("RA_TWILIO_AUTH_TOKEN"),
			phoneNumber:  os.Getenv("RA_TWILIO_PHONE_NUMBER"),
			twilioNumber: os.Getenv("RA_TWILIO_TWILIO_NUMBER"),
		}
	}

}

func (s *Service) GetAction(name string) Action {
	action, ok := s.actions[name]
	if !ok {
		panic("Action has not been registered!")
	}
	return action
}
