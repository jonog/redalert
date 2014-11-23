package main

type Service struct {
	servers []*Server
	alerts  map[string]Alert
}

func (s *Service) Start() {
	stopTimer := make(chan bool)
	for _, server := range s.servers {
		go server.Monitor()
	}
	<-stopTimer
}

func main() {

	service := new(Service)
	service.SetupAlerts()

	config, err := ReadConfigFile()
	if err != nil {
		panic("Missing or invalid config")
	}

	for _, sc := range config.Servers {
		service.AddServer(sc.Name, sc.Address, sc.Interval, sc.Alerts)
	}

	service.Start()

}
