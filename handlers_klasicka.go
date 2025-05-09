package main

import (
	"net/http"
	"time"

	"github.com/coreos/go-log/log"
	"github.com/go-chi/chi/v5"
)

func klasickaRouter() *chi.Mux {
	r := newGameRouter("klasicka")
	r.Get("/", auth(klasickaIndexGet))
	r.Post("/", auth(klasickaIndexPost))
	return r
}

type klasickaIndexData struct {
	GeneralData
}

func klasickaIndexGet(w http.ResponseWriter, r *http.Request) {
	team := server.state.GetTeam(getUser(r)) // Uncomment if team data is needed

	if team != nil && team.Klasicka.Completed {
		data := klasickaIndexData{
			GeneralData: getGeneralData("Hroší opera - Hacked", w, r),
		}
		executeTemplate(w, "klasickaHacked", data)
	} else {
		data := klasickaIndexData{
			GeneralData: getGeneralData("Hroší opera", w, r),
		}
		executeTemplate(w, "klasickaIndex", data)
	}
}


func klasickaIndexPost(w http.ResponseWriter, r *http.Request) {
	defer http.Redirect(w, r, "/", http.StatusSeeOther)

	if err := r.ParseForm(); err != nil {
		setFlashMessage(w, r, messageError, "Cannot parse form")
		return
	}

	server.state.Lock()
	defer server.state.Unlock()

	team := server.state.GetTeam(getUser(r))
	if team != nil && (team.Klasicka.Completed) {
		return
	}

	
	cislo := r.PostFormValue("cislo")
	heslo := r.PostFormValue("heslo")
	log.Infof("[Klasicka - %s] Trying login: '%s' and pswd: %d", team.Login, cislo, heslo)
	if cislo == "78136" && heslo == "amadeus69" {
		log.Infof("[Klasicka - %s] Login passed!", team.Login)
		team.Klasicka.Completed = true
		team.Klasicka.CompletedTime = time.Now()
	}
	// slog := slog.With("team", team.Login)

	team.Klasicka.Tries++
	team.Klasicka.LastTry = time.Now()
	defer server.state.Save()
}
