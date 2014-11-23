package main

import (
	"fmt"
	"os"
)

type StandardError struct{}

func (a StandardError) Trigger(event *Event) error {
	fmt.Fprintln(os.Stderr, event.ShortMessage())
	return nil
}
