package web

import (
	"net/http"

	"github.com/gorilla/mux"
)

func checkDisableHandler(c *appCtx, w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id := vars["check_id"]

	check, err := c.service.CheckByID(id)
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	if !check.Data.Enabled {
		http.Error(w, "Check is already disabled", http.StatusPreconditionFailed)
		return
	}

	check.Stop()

	w.Write([]byte(`OK`))
}

func checkEnableHandler(c *appCtx, w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id := vars["check_id"]

	check, err := c.service.CheckByID(id)
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	if check.Data.Enabled {
		http.Error(w, "Check is already disabled", http.StatusPreconditionFailed)
		return
	}

	go check.Start()

	w.Write([]byte(`OK`))
}

func checkTriggerHandler(c *appCtx, w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id := vars["check_id"]

	check, err := c.service.CheckByID(id)
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	if !check.Data.Enabled {
		http.Error(w, "Check is not enabled", http.StatusPreconditionFailed)
		return
	}

	check.Trigger()

	w.Write([]byte(`OK`))
}
