package main

func main() {

	servers := []*Server{
		NewServer("Server 1", "http://server1.com/healthcheck", 3, []string{"console_message"}),
		NewServer("Server 2", "http://server2.com/healthcheck", 3, []string{"console_message"}),
		NewServer("Server 3", "http://server3.com/healthcheck", 3, []string{"console_message"}),
	}

	for _, server := range servers {
		go server.Monitor()
	}

	stopTimer := make(chan bool)
	<-stopTimer

}
