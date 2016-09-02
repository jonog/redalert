package config

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"time"

	"github.com/jonog/redalert/checks"
	"github.com/jonog/redalert/notifiers"
)

type FileStore struct {
	filename string
	data     FileStoreData
}

type FileStoreData struct {
	Checks        []checks.Config    `json:"checks"`
	Notifications []notifiers.Config `json:"notifications"`
	Preferences   Preferences        `json:"preferences"`
}

func NewFileStore(filename string) (*FileStore, error) {
	config := &FileStore{filename: filename}
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

	err = config.write()
	if err != nil {
		return nil, err
	}

	return config, nil
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

var idLetters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func generateID(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = idLetters[rand.Intn(len(idLetters))]
	}
	return string(b)
}

func (f *FileStore) read() error {
	file, err := ioutil.ReadFile(f.filename)
	if err != nil {
		return err
	}
	var data FileStoreData
	err = json.Unmarshal(file, &data)
	if err != nil {
		return err
	}
	f.data = data
	return nil
}

func (f *FileStore) write() error {
	b, err := json.MarshalIndent(f.data, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(f.filename, b, 0644)
}

func (f *FileStore) Notifications() ([]notifiers.Config, error) {
	return f.data.Notifications, nil
}

func (f *FileStore) Checks() ([]checks.Config, error) {
	return f.data.Checks, nil
}

func (f *FileStore) Preferences() (Preferences, error) {
	return f.data.Preferences, nil
}
