package notifiers

import (
	"log"
	"os"
)

type StandardError struct {
	name string
	log  *log.Logger
}

func NewStandardError() StandardError {
	return StandardError{
		name: "stderr",
		log:  log.New(os.Stderr, "", log.Ldate|log.Ltime),
	}
}

func (a StandardError) Name() string {
	return a.name
}

func (a StandardError) Trigger(msg Message) error {
	a.log.Println(msg.ShortMessage())
	return nil
}
