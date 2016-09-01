package config

import (
	"github.com/jonog/redalert/checks"
	"github.com/jonog/redalert/notifiers"
)

type Store interface {
	Notifications() ([]notifiers.Config, error)
	Checks() ([]checks.Config, error)
}

type writableStore interface {
	Store
	createOrUpdateNotificationRecord(notifier notifiers.Config) error
	createOrUpdateCheckRecord(check checks.Config) error
}

func Copy(srcStore Store, destStore writableStore) error {

	notifications, err := srcStore.Notifications()
	if err != nil {
		return err
	}
	for _, n := range notifications {
		err = destStore.createOrUpdateNotificationRecord(n)
		if err != nil {
			return err
		}
	}

	checks, err := srcStore.Checks()
	if err != nil {
		return err
	}
	for _, c := range checks {
		err = destStore.createOrUpdateCheckRecord(c)
		if err != nil {
			return err
		}
	}

	return nil
}
