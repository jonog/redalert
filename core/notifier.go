package core

import (
	"errors"

	"github.com/jonog/redalert/notifiers"
)

func (s *Service) RegisterNotification(config notifiers.Config) error {
	notifier, err := NewNotifier(config)
	if err != nil {
		return err
	}
	s.Notifiers[notifier.Name()] = notifier
	return nil
}

func NewNotifier(config notifiers.Config) (Notifier, error) {

	var notifier Notifier
	var err error

	switch config.Type {
	case "gmail":
		notifier, err = notifiers.NewGmailNotifier(config)
	case "slack":
		notifier, err = notifiers.NewSlackNotifier(config)
	case "twilio":
		notifier, err = notifiers.NewTwilioNotifier(config)
	default:
		return nil, errors.New("redalert: unknown notifier")
	}

	return notifier, err

}
