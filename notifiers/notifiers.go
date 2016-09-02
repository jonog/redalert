package notifiers

import (
	"errors"

	"github.com/jonog/redalert/events"
)

type Notifier interface {
	Notify(Message) error
	Name() string
}

type Message struct {
	DefaultMessage string
	Event          *events.Event
}

type Config struct {
	ID     string            `json:"id"`
	Name   string            `json:"name"`
	Type   string            `json:"type"`
	Config map[string]string `json:"config"`
}

var registry = make(map[string]func(Config) (Notifier, error))

func registerNotifier(name string, constructorFn func(Config) (Notifier, error)) {
	registry[name] = constructorFn
}

func New(config Config) (Notifier, error) {
	notifierFn, ok := registry[config.Type]
	if !ok {
		return nil, errors.New("notifiers: notifier unavailable: " + config.Type)
	}
	return notifierFn(config)
}
