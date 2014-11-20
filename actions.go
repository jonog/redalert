package main

type Action interface {
	Send(*Server) error
}

type ConsoleMessage struct{}

func (a ConsoleMessage) Send(server *Server) error {
	server.log.Println("Time for action!")
	return nil
}

// TODO
type SlackWebhook struct{}
type Email struct{}
type SMS struct{}
type ExecuteCommand struct{}
