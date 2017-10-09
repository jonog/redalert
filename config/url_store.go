package config

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/jonog/redalert/checks"
	"github.com/jonog/redalert/notifiers"
)

type URLStore struct {
	URL  string
	data URLStoreData
}

type URLStoreData struct {
	Checks        []checks.Config    `json:"checks"`
	Notifications []notifiers.Config `json:"notifications"`
	Preferences   Preferences        `json:"preferences"`
}

func NewURLStore(URL string) (*URLStore, error) {
	config := &URLStore{URL: URL}
	err := config.read()
	if err != nil {
		return nil, err
	}

	// create check ID if not present
	for i := range config.data.Checks {
		if config.data.Checks[i].ID == "" {
			config.data.Checks[i].ID = generateID(8)
		}
	}

	// create notification ID if not present
	for i := range config.data.Notifications {
		if config.data.Notifications[i].ID == "" {
			config.data.Notifications[i].ID = generateID(8)
		}
	}

	return config, nil
}

func (u *URLStore) read() error {
	resp, err := http.Get(u.URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var data URLStoreData
	err = json.Unmarshal(body, &data)
	if err != nil {
		return err
	}
	u.data = data
	return nil
}

func (u *URLStore) Notifications() ([]notifiers.Config, error) {
	return u.data.Notifications, nil
}

func (u *URLStore) Checks() ([]checks.Config, error) {
	return u.data.Checks, nil
}

func (u *URLStore) Preferences() (Preferences, error) {
	return u.data.Preferences, nil
}
