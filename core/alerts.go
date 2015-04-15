package core

type Alert interface {
	Trigger(*Event) error
	Name() string
}

func (s *Service) GetAlert(name string) Alert {
	alert, ok := s.Alerts[name]
	if !ok {
		panic("Alert has not been registered!")
	}
	return alert
}
