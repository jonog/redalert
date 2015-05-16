package notifiers

type Config struct {
	Name   string
	Type   string
	Config map[string]string
}

type Message interface {
	ShortMessage() string
}
