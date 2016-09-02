package config

import (
	"github.com/jonog/redalert/checks"
	"github.com/jonog/redalert/notifiers"
)

type Store interface {
	Notifications() ([]notifiers.Config, error)
	Checks() ([]checks.Config, error)
	Preferences() (Preferences, error)
}

type writableStore interface {
	Store
	createOrUpdateNotification(notifier notifiers.Config) error
	createOrUpdateCheck(check checks.Config) error
	updatePreferences(preferences Preferences) error
}

func Copy(srcStore Store, destStore writableStore) error {

	notifications, err := srcStore.Notifications()
	if err != nil {
		return err
	}
	for _, n := range notifications {
		err = destStore.createOrUpdateNotification(n)
		if err != nil {
			return err
		}
	}

	checks, err := srcStore.Checks()
	if err != nil {
		return err
	}
	for _, c := range checks {
		err = destStore.createOrUpdateCheck(c)
		if err != nil {
			return err
		}
	}

	preferences, err := srcStore.Preferences()
	if err != nil {
		return err
	}
	return destStore.updatePreferences(preferences)
}
