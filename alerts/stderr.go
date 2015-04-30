package alerts

import (
	"log"
	"os"

	"github.com/jonog/redalert/core"
)

type StandardError struct {
	log *log.Logger
}

func NewStandardError() StandardError {
	return StandardError{
		log: log.New(os.Stderr, "", log.Ldate|log.Ltime),
	}
}

func (a StandardError) Name() string {
	return "StandardError"
}

func (a StandardError) Trigger(event *core.Event) error {
	a.log.Println(event.ShortMessage())
	event.Check.Log.Println("Stderr alert successfully triggered.")
	return nil
}
