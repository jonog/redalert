package main

import (
	"log"
	"os"
	"strings"
	"time"
)

type Alert interface {
	Trigger(*Event) error
}

type Event struct {
	server *Server
	time   time.Time
}

func (e *Event) ShortMessage() string {
	return strings.Join([]string{"Uhoh,", e.server.name, "not responding. Failed ping to", e.server.address}, " ")
}

func (s *Service) SetupAlerts() {

	logger := log.New(os.Stdout, "Setup ", log.Ldate|log.Ltime)

	s.alerts = make(map[string]Alert)

	s.alerts["stderr"] = StandardError{}

	if os.Getenv("RA_SLACK_URL") == "" {
		logger.Println("Slack is not configured")
	} else {
		s.alerts["slack"] = SlackWebhook{url: os.Getenv("RA_SLACK_URL")}
	}

	if os.Getenv("RA_GMAIL_USER") == "" || os.Getenv("RA_GMAIL_PASS") == "" || os.Getenv("RA_GMAIL_NOTIFICATION_ADDRESS") == "" {
		logger.Println("Email is not configured")
	} else {
		s.alerts["email"] = Email{
			user:                os.Getenv("RA_GMAIL_USER"),
			pass:                os.Getenv("RA_GMAIL_PASS"),
			notificationAddress: os.Getenv("RA_GMAIL_NOTIFICATION_ADDRESS"),
		}
	}

	if os.Getenv("RA_TWILIO_ACCOUNT_SID") == "" || os.Getenv("RA_TWILIO_AUTH_TOKEN") == "" || os.Getenv("RA_TWILIO_PHONE_NUMBER") == "" || os.Getenv("RA_TWILIO_TWILIO_NUMBER") == "" {
		logger.Println("SMS is not configured")
	} else {
		s.alerts["sms"] = SMS{
			accountSid:   os.Getenv("RA_TWILIO_ACCOUNT_SID"),
			authToken:    os.Getenv("RA_TWILIO_AUTH_TOKEN"),
			phoneNumber:  os.Getenv("RA_TWILIO_PHONE_NUMBER"),
			twilioNumber: os.Getenv("RA_TWILIO_TWILIO_NUMBER"),
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
