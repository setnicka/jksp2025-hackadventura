package main

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

func metalRouter() *chi.Mux {
	r := newGameRouter("metal")
	r.Get("/", auth(metalIndexGet))
	r.Post("/", auth(metalIndexPost))
	return r
}

func metalIndexGet(w http.ResponseWriter, r *http.Request) {
	team := server.state.GetTeam(getUser(r))

	if team != nil && team.Metal.Completed {
		data := getGeneralData("Hotel", w, r)
		executeTemplate(w, "metalHacked", data)

	} else {
		data := getGeneralData("Hotel", w, r)
		executeTemplate(w, "metalIndex", data)
	}
}

func metalIndexPost(w http.ResponseWriter, r *http.Request) {
	defer http.Redirect(w, r, "/", http.StatusSeeOther)

	if err := r.ParseForm(); err != nil {
		setFlashMessage(w, r, messageError, "Cannot parse form")
		return
	}

	server.state.Lock()
	defer server.state.Unlock()

	team := server.state.GetTeam(getUser(r))
	if team != nil && (team.Metal.Completed) {
		return
	}

	// slog := slog.With("team", team.Login)

	team.Metal.Tries++
	team.Metal.LastTry = time.Now()
	defer server.state.Save()

	////////////////////////////////////////////////////////////////

	// slog.Info("[Hotel] completed")
	// // Everything completed
	// team.Hotel.Completed = true
	// team.Hotel.CompletedTime = time.Now()
}
