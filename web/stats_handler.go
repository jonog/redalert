package web

import (
	"net/http"

	"github.com/jonog/redalert/events"
)

type CheckResponse struct {
	Id     string          `json:"id"`
	Name   string          `json:"name"`
	Events []*events.Event `json:"events"`
}

func statsHandler(c *appCtx, w http.ResponseWriter, r *http.Request) {

	checks := c.service.Checks()
	displayChecks := make([]CheckResponse, len(checks))

	for idx, check := range checks {
		events, err := check.Store.GetRecent()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		displayChecks[idx] = CheckResponse{
			Id:     check.Id,
			Name:   check.Name,
			Events: events,
		}
	}

	Respond(w, displayChecks, 200)
}
