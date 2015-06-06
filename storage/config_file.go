package storage

import (
	"encoding/json"
	"io/ioutil"

	"github.com/jonog/redalert/checks"
	"github.com/jonog/redalert/notifiers"
)

type ConfigFile struct {
	filename string
	data     ConfigFileData
}

type ConfigFileData struct {
	Checks        []checks.Config    `json:"checks"`
	Notifications []notifiers.Config `json:"notifications"`
}

func NewConfigFile(filename string) (*ConfigFile, error) {
	config := &ConfigFile{filename: filename}
	err := config.read(filename)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func (f *ConfigFile) read(filename string) error {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	var data ConfigFileData
	err = json.Unmarshal(file, &data)
	if err != nil {
		return err
	}
	f.data = data
	return nil
}

func (f *ConfigFile) Notifications() ([]notifiers.Config, error) {
	return f.data.Notifications, nil
}

func (f *ConfigFile) Checks() ([]checks.Config, error) {
	return f.data.Checks, nil
}
