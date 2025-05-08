package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func klasickaRouter() *chi.Mux {
	r := newGameRouter("klasicka")
	r.Get("/", klasickaIndexGet)
	return r
}

type klasickaIndexData struct {
	GeneralData
}

func klasickaIndexGet(w http.ResponseWriter, r *http.Request) {
	// team := server.state.GetTeam(getUser(r)) // Uncomment if team data is needed

	data := klasickaIndexData{
		GeneralData: getGeneralData("Hroší opera", w, r),
	}
	executeTemplate(w, "klasickaIndex", data)
}
