package web

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/GeertJohan/go.rice"
	"github.com/jonog/redalert/core"
	"github.com/rs/cors"
)

func Run(service *core.Service, port string) {

	context := &appCtx{
		service: service,
	}

	box := rice.MustFindBox("static")
	fs := http.FileServer(box.HTTPBox())

	mux := http.NewServeMux()

	mux.Handle("/static/", http.StripPrefix("/static/", fs))
	mux.Handle("/", appHandler{context, dashboardHandler})
	mux.Handle("/api/put", appHandler{context, metricsReceiverHandler})

	mux.Handle("/v1/stats", appHandler{context, statsHandler})

	mux.HandleFunc("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	handler := cors.Default().Handler(mux)
	err := http.ListenAndServe(":8888", handler)
	if err != nil {
		log.Fatal(err)
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

func Respond(w http.ResponseWriter, data interface{}, code int) {
	response, err := json.Marshal(data)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
