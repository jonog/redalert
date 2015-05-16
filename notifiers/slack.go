package notifiers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
)

type SlackWebhook struct {
	name      string
	url       string
	channel   string
	username  string
	iconEmoji string
}

func NewSlackNotifier(config Config) (SlackWebhook, error) {

	if config.Type != "slack" {
		return SlackWebhook{}, errors.New("slack: invalid config type")
	}

	if config.Config["webhook_url"] == "" {
		return SlackWebhook{}, errors.New("slack: invalid webhook_url")
	}

	return SlackWebhook{
		name:      config.Name,
		url:       config.Config["webhook_url"],
		channel:   config.Config["channel"],
		username:  config.Config["username"],
		iconEmoji: config.Config["icon_emoji"],
	}, nil
}

func (a SlackWebhook) Name() string {
	return a.name
}

func (a SlackWebhook) Trigger(msg Message) error {

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
		Text:      msg.ShortMessage(),
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

	return nil
}

type SlackPayload struct {
	Channel   string `json:"channel"`
	Username  string `json:"username,omitempty"`
	Text      string `json:"text"`
	Parse     string `json:"parse"`
	IconEmoji string `json:"icon_emoji,omitempty"`
}
