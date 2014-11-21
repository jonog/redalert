package main

import (
	"log"
	"os"
)

type Action interface {
	Send(*Server) error
}

func (s *Service) SetupActions() {

	s.actions = make(map[string]Action)

	s.actions["console"] = ConsoleMessage{}

	if os.Getenv("RA_SLACK_URL") == "" {
		log.Println("Slack is not configured")
	} else {
		s.actions["slack"] = SlackWebhook{url: os.Getenv("RA_SLACK_URL")}
	}

	if os.Getenv("RA_GMAIL_USER") == "" || os.Getenv("RA_GMAIL_PASS") == "" || os.Getenv("RA_GMAIL_NOTIFICATION_ADDRESS") == "" {
		log.Println("Email is not configured")
	} else {
		s.actions["email"] = Email{
			user:                os.Getenv("RA_GMAIL_USER"),
			pass:                os.Getenv("RA_GMAIL_PASS"),
			notificationAddress: os.Getenv("RA_GMAIL_NOTIFICATION_ADDRESS"),
		}
	}

}

func (s *Service) GetAction(name string) Action {

	action, ok := s.actions[name]
	if !ok {
		panic("Unknown action!")
	}

	return action

}

type ConsoleMessage struct{}

func (a ConsoleMessage) Send(server *Server) error {
	server.log.Println("Time for action!")
	return nil
}

type SMS struct{}
type ExecuteCommand struct{}
