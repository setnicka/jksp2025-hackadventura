package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func technoRouter() *chi.Mux {
	r := newGameRouter("techno")
	r.Get("/", technoIndexGet)
	return r
}

type technoIndexData struct {
	GeneralData
}

func technoIndexGet(w http.ResponseWriter, r *http.Request) {
	// team := server.state.GetTeam(getUser(r)) // Uncomment if team data is needed

	data := technoIndexData{
		GeneralData: getGeneralData("Techno Trosky - Vítejte v Matrixu", w, r),
	}
	executeTemplate(w, "technoIndex", data)
}