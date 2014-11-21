package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
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

type SlackWebhook struct {
	url string
}

func (a SlackWebhook) Send(server *Server) error {
	message := SlackPayload{
		Channel:   "#general",
		Username:  "redalert",
		Text:      "Uhoh, " + server.name + " has been nuked!!!",
		Parse:     "full",
		IconEmoji: ":rocket:",
	}

	buf, err := json.Marshal(message)
	if err != nil {
		return err
	}

	resp, err := http.Post(a.url, "application/json", bytes.NewBuffer(buf))
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New("Not OK")
	}
	return nil
}

type SlackPayload struct {
	Channel   string `json:"channel"`
	Username  string `json:"username,omitempty"`
	Text      string `json:"text"`
	Parse     string `json:"parse"`
	IconEmoji string `json:"icon_emoji,omitempty"`
}

// TODO
type Email struct{}
type SMS struct{}
type ExecuteCommand struct{}
