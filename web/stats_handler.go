package web

import (
	"net/http"

	"github.com/jonog/redalert/events"
)

type checkPublic struct {
	ID     string          `json:"id"`
	Name   string          `json:"name"`
	Events []*events.Event `json:"events"`
}

func statsHandler(c *appCtx, w http.ResponseWriter, r *http.Request) {

	checks := c.service.Checks()
	publicChecks := make([]checkPublic, len(checks))

	for idx, check := range checks {
		events, err := check.Store.GetRecent()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		publicChecks[idx] = checkPublic{
			ID:     check.Data.ID,
			Name:   check.Data.Name,
			Events: events,
		}
	}

	Respond(w, publicChecks, 200)
}
