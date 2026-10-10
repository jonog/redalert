package config

import (
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/jonog/redalert/checks"
	"github.com/jonog/redalert/notifiers"
)

type FileStore struct {
	filename string
	data     FileStoreData
	mu       sync.Mutex
}

func (f *FileStore) Filename() string { return f.filename }

// AppendChecks atomically appends checks while preserving all other file data.
func (f *FileStore) AppendChecks(additions []checks.Config) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	file, err := ioutil.ReadFile(f.filename)
	if err != nil {
		return err
	}
	var data FileStoreData
	if err = decodeConfig(file, formatForPath(f.filename), &data); err != nil {
		return err
	}
	data.Checks = append(data.Checks, additions...)
	b, err := encodeConfig(data, formatForPath(f.filename))
	if err != nil {
		return err
	}
	dir := filepath.Dir(f.filename)
	tmp, err := ioutil.TempFile(dir, ".redalert-config-*")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer os.Remove(name)
	if _, err = tmp.Write(b); err != nil {
		tmp.Close()
		return err
	}
	if err = tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err = tmp.Close(); err != nil {
		return err
	}
	if err = os.Rename(name, f.filename); err != nil {
		return err
	}
	f.data = data
	return nil
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
	err = decodeConfig(file, formatForPath(f.filename), &data)
	if err != nil {
		return err
	}
	f.data = data
	return nil
}

func (f *FileStore) write() error {
	b, err := encodeConfig(f.data, formatForPath(f.filename))
	if err != nil {
		return err
	}
	return ioutil.WriteFile(f.filename, b, 0644)
}

func (f *FileStore) Notifications() ([]notifiers.Config, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.data.Notifications, nil
}

func (f *FileStore) Checks() ([]checks.Config, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.data.Checks, nil
}

func (f *FileStore) Preferences() (Preferences, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.data.Preferences, nil
}
