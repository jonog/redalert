package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"text/template"

	"github.com/GeertJohan/go.rice"
)

type Service struct {
	servers []*Server
	alerts  map[string]Alert
	wg      sync.WaitGroup
}

func (s *Service) Start() {

	// use this to keep the service running, even if no monitoring is occuring
	s.wg.Add(1)

	for _, server := range s.servers {
		go server.Monitor()
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for _ = range c {
			s.wg.Done()
		}
	}()

}

type DashboardInfo struct {
	Servers []*Server
}

func dashboardHandler(c *appCtx, w http.ResponseWriter, r *http.Request) {

	templateBox, err := rice.FindBox("templates")
	if err != nil {
		log.Println(err.Error())
		http.Error(w, http.StatusText(500), 500)
		return
	}

	templateString, err := templateBox.String("dash.html")
	if err != nil {
		log.Println(err.Error())
		http.Error(w, http.StatusText(500), 500)
		return
	}

	tmplMessage, err := template.New("dash").Parse(templateString)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, http.StatusText(500), 500)
		return
	}

	if err != nil {
		log.Println(err.Error())
		http.Error(w, http.StatusText(500), 500)
		return
	}

	info := &DashboardInfo{Servers: c.service.servers}

	if err := tmplMessage.Execute(w, info); err != nil {
		log.Println(err.Error())
		http.Error(w, http.StatusText(500), 500)
	}
}

type appCtx struct {
	service *Service
}

type appHandler struct {
	*appCtx
	h func(*appCtx, http.ResponseWriter, *http.Request)
}

func (ah appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ah.h(ah.appCtx, w, r)
}

func getPort() string {
	if os.Getenv("RA_PORT") == "" {
		return "8888"
	} else {
		return os.Getenv("RA_PORT")
	}
}

func main() {

	service := new(Service)

	config, err := ReadConfigFile()
	if err != nil {
		panic("Missing or invalid config")
	}

	service.SetupAlerts(config)

	for _, sc := range config.Servers {
		service.AddServer(sc.Name, sc.Address, sc.Interval, sc.Alerts)
	}

	service.Start()
	context := &appCtx{
		service: service,
	}

	go func() {
		box := rice.MustFindBox("static")
		fs := http.FileServer(box.HTTPBox())
		http.Handle("/static/", http.StripPrefix("/static/", fs))
		http.Handle("/", appHandler{context, dashboardHandler})

		port := getPort()
		fmt.Println("Listening on port ", port, " ...")
		http.ListenAndServe(":"+port, nil)
	}()

	service.wg.Wait()

}
