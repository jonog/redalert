package core

import "github.com/jonog/redalert/notifiers"

func (s *Service) RegisterNotifier(config notifiers.Config) error {
	notifier, err := notifiers.New(config)
	if err != nil {
		return err
	}
	s.Notifiers[notifier.Name()] = notifier
	return nil
}
