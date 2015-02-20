package main

import (
	"log"
	"os"
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

func (a StandardError) Trigger(event *Event) error {
	a.log.Println(event.ShortMessage())
	event.Server.log.Println(white, "Stderr alert successfully triggered.", reset)
	return nil
}
