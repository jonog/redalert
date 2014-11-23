package main

import (
	"fmt"
	"os"
)

type StandardError struct{}

func (a StandardError) Trigger(event *Event) error {
	fmt.Fprintln(os.Stderr, event.ShortMessage())
	event.server.log.Println(white, "Stderr alert successfully triggered.", reset)
	return nil
}
