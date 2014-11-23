package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
)

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
