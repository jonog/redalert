package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
)

type SlackWebhook struct {
	url       string
	channel   string
	username  string
	iconEmoji string
}

func (a SlackWebhook) Name() string {
	return "SlackWebhook"
}

func (a SlackWebhook) Trigger(event *Event) error {

	var payloadChannel string
	var payloadUsername string
	var payloadIconEmoji string

	if a.channel == "" {
		payloadChannel = "#general"
	} else {
		payloadChannel = a.channel
	}

	if a.username == "" {
		payloadUsername = "redalert"
	} else {
		payloadUsername = a.username
	}

	if a.iconEmoji == "" {
		payloadIconEmoji = ":rocket:"
	} else {
		payloadIconEmoji = a.iconEmoji
	}

	message := SlackPayload{
		Channel:   payloadChannel,
		Username:  payloadUsername,
		Text:      event.ShortMessage(),
		Parse:     "full",
		IconEmoji: payloadIconEmoji,
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

	event.Server.log.Println(white, "Slack alert successfully triggered.", reset)
	return nil
}

type SlackPayload struct {
	Channel   string `json:"channel"`
	Username  string `json:"username,omitempty"`
	Text      string `json:"text"`
	Parse     string `json:"parse"`
	IconEmoji string `json:"icon_emoji,omitempty"`
}
