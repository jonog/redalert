package notifiers

import (
	"log"
	"os"
)

func init() {
	registerNotifier("stderr", NewStandardError)
}

type StandardError struct {
	name string
	log  *log.Logger
}

var NewStandardError = func(config Config) (Notifier, error) {
	return Notifier(StandardError{
		name: "stderr",
		log:  log.New(os.Stderr, "", log.Ldate|log.Ltime),
	}), nil
}

func (a StandardError) Name() string {
	return a.name
}

func (a StandardError) Notify(msg Message) error {
	a.log.Println(msg.DefaultMessage)
	return nil
}
