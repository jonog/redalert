package notifiers

import (
	"errors"
)

type Notifier interface {
	Notify(Message) error
	Name() string
}

type Message interface {
	ShortMessage() string
}

/////////////////
// Initialisation
/////////////////

type Config struct {
	Name   string
	Type   string
	Config map[string]string
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
