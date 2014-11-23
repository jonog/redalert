package main

import (
	"fmt"
	"os"
)

type StandardError struct{}

func (a StandardError) Send(server *Server) error {
	server.log.Println()
	fmt.Fprintln(os.Stderr, "Uhoh, "+server.name+" has been nuked!!!")
	return nil
}
