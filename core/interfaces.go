package core

type Alert interface {
	Trigger(*Event) error
	Name() string
}

type Checker interface {
	Check() (map[string]float64, error)
	RedAlertMessage() string
	GreenAlertMessage() string
}
