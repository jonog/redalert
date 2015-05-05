package web

import (
	"fmt"
	"net/http"

	"github.com/GeertJohan/go.rice"
	"github.com/jonog/redalert/core"
)

func Run(service *core.Service, port string) {

	context := &appCtx{
		service: service,
	}

	box := rice.MustFindBox("static")
	fs := http.FileServer(box.HTTPBox())

	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.Handle("/", appHandler{context, dashboardHandler})
	http.Handle("/api/put", appHandler{context, metricsReceiverHandler})

	fmt.Println("Listening on port ", port, " ...")
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		panic(err)
	}
}

type appCtx struct {
	service *core.Service
}

type appHandler struct {
	*appCtx
	h func(*appCtx, http.ResponseWriter, *http.Request)
}

func (ah appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ah.h(ah.appCtx, w, r)
}
