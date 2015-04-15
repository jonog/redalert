package web

import (
	"log"
	"net/http"
	"text/template"

	"github.com/GeertJohan/go.rice"
	"github.com/jonog/redalert/core"
)

type DashboardInfo struct {
	Servers []*core.Server
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

	info := &DashboardInfo{Servers: c.service.Servers}

	if err := tmplMessage.Execute(w, info); err != nil {
		log.Println(err.Error())
		http.Error(w, http.StatusText(500), 500)
	}
}
