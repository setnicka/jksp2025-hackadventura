package main

import (
	"net/http"
	"time"

	"github.com/coreos/go-log/log"
	"github.com/go-chi/chi/v5"
)

func metalRouter() *chi.Mux {
	r := newGameRouter("metal")
	r.Get("/", auth(metalIndexGet))
	r.Post("/", auth(metalIndexPost))
	return r
}

const (
	metalLogin    = "zahradnik"
	metalPassword = "trubkyvmoshpitu"
)

func metalIndexGet(w http.ResponseWriter, r *http.Request) {
	team := server.state.GetTeam(getUser(r))

	if team != nil && team.Metal.Completed {
		data := getGeneralData("Přístup povolen", w, r)
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
	login := r.FormValue("login")
	password := r.FormValue("password")
	// slog := slog.With("team", team.Login)

	team.Metal.Tries++
	team.Metal.LastTry = time.Now()
	defer server.state.Save()

	// log that the team is trying to log in
	log.Infof("[Metal - %s] Trying login '%s' and password '%s'", team.Login, login, password)
	if login == metalLogin && password == metalPassword {
		log.Infof("[Metal - %s] Completed", team.Login)
		team.Metal.Completed = true
		team.Metal.CompletedTime = time.Now()
		server.state.Save()
		setFlashMessage(w, r, messageOk, "Login successful")
		return
	} else {
		log.Warningf("[Metal - %s] Failed login attempt. Provided Login: '%s', Expected Login: '%s', Provided Password: '%s', Expected Password: '%s'", team.Login, login, metalLogin, password, metalPassword)
		setFlashMessage(w, r, messageError, "Wrong login or password")
		return
	}
}
