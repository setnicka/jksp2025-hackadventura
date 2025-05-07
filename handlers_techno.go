package main

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/coreos/go-log/log"
	
)

func technoRouter() *chi.Mux {
	r := newGameRouter("techno")
	r.Get("/", auth(technoIndexGet))
	r.Post("/", technoIndexPost)
	r.Get(technoFinalURL, auth(technoIntranetGet))
	return r
}

const (
	technoPassword = "12345678" // TODO: change before the game
	TechnoName     = "Tech Troska" // display in frontend
	technoFinalURL = "/techno_trosky"
)

type technoIndexData struct {
	GeneralData
	Completed bool
	Name      string
}

func technoIndexGet(w http.ResponseWriter, r *http.Request) {
	team := server.state.GetTeam(getUser(r))

	data := technoIndexData{
		GeneralData: getGeneralData("Techno Trosky - Vítejte v Matrixu", w, r),
		Completed:   team.Techno.Completed,
		Name:        TechnoName,
	}
	executeTemplate(w, "technoIndex", data)
}


func technoIndexPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		setFlashMessage(w, r, messageError, "Cannot parse login form")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	server.state.Lock()
	defer server.state.Unlock()

	team := server.state.GetTeam(getUser(r))
	if team != nil && (team.Techno.Completed) {
		http.Redirect(w, r, technoFinalURL, http.StatusSeeOther)
		return
	}

	team.Techno.Tries++

	login := r.PostFormValue("login")
	password := r.PostFormValue("password")
	log.Infof("[Techno - %s] Trying login '%s' and password '%s'", team.Login, login, password)

	// Perform case-sensitive password comparison
	if password == technoPassword {
		log.Infof("[Techno - %s] Completed", team.Login)
		// Everything completed
		team.Techno.Completed = true
		team.Techno.CompletedTime = time.Now()
		// Save state before redirecting (optional but safer)
		server.state.Save()
		setFlashMessage(w, r, messageOk, "Přihlášení bylo úspěšné, vítejte v tajném doupěti Technarů")
		http.Redirect(w, r, technoFinalURL, http.StatusSeeOther)
	} else {
		// Add more detailed logging for failed attempts
		log.Warningf("[Techno - %s] Failed login attempt. Provided Login: '%s', Provided Password: '%s', Expected Password: '%s'", team.Login, login, password, technoPassword)
		setFlashMessage(w, r, messageError, "Nesprávné heslo")
		// Explicit redirect on failure (already present)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func technoIntranetGet(w http.ResponseWriter, r *http.Request) {
	team := server.state.GetTeam(getUser(r))
	if team == nil || !team.Techno.Completed {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	data := technoIndexData{
		GeneralData: getGeneralData("Techno – ", w, r),
		Completed:   team.Techno.Completed,
		Name:        TechnoName,
	}

	executeTemplate(w, "technoIntranet", data)
}