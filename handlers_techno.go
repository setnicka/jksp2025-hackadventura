package main

import (
	"net/http"
	"strings"
	"time"

	"github.com/coreos/go-log/log"
	"github.com/go-chi/chi/v5"
)

func technoRouter() *chi.Mux {
	r := newGameRouter("techno")
	r.Get("/", auth(technoIndexGet))
	r.Post("/", auth(technoIndexPost)) // Added auth here as well for consistency, user must exist
	r.Get(technoFinalURL, auth(technoIntranetGet))
	return r
}

const (
	technoPassword = "vlakvopere"    
	TechnoName     = "Tech Troska"   // display in frontend
	technoFinalURL = "/techno_trosky" 
	technoLogin    = "zahradnik" 
)

type technoIndexData struct {
	GeneralData
	Completed bool
	Name      string
}

func technoIndexGet(w http.ResponseWriter, r *http.Request) {
	team := server.state.GetTeam(getUser(r))
	if team != nil && team.Techno.Completed {
		http.Redirect(w, r, technoFinalURL, http.StatusSeeOther) // Uses corrected technoFinalURL
		return
	}

	data := technoIndexData{
		GeneralData: getGeneralData("Techno Trosky - Vítejte v Matrixu", w, r),
		Name:        TechnoName,
		Completed:   team != nil && team.Techno.Completed,
	}
	executeTemplate(w, "technoIndex", data)
}

func technoIndexPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		setFlashMessage(w, r, messageError, "Cannot parse login form")
		http.Redirect(w, r, "/", http.StatusSeeOther) // Relative to /techno
		return
	}

	server.state.Lock()
	defer server.state.Unlock()

	team := server.state.GetTeam(getUser(r))
	if team == nil {
		log.Warningf("[Techno] Post attempt by non-existent or non-authenticated user")
		http.Redirect(w, r, "/start-hry", http.StatusSeeOther)
		return
	}

	if team.Techno.Completed {
		http.Redirect(w, r, technoFinalURL, http.StatusSeeOther) // Uses corrected technoFinalURL
		return
	}

	team.Techno.Tries++
	team.Techno.LastTry = time.Now()

	login := r.PostFormValue("login")
	password := r.PostFormValue("password")
	// Remove diacritics from login and password
	login = strings.ToLower(strings.TrimSpace(login))
	password = strings.ToLower(strings.TrimSpace(password))

	log.Infof("[Techno - %s] Trying login '%s' and password '%s'", team.Login, login, password)

	if login == technoLogin && password == technoPassword {
		log.Infof("[Techno - %s] Completed", team.Login)
		team.Techno.Completed = true
		team.Techno.CompletedTime = time.Now()
		server.state.Save()
		setFlashMessage(w, r, messageOk, "Přihlášení bylo úspěšné, vítejte v tajném doupěti Technarů")
		http.Redirect(w, r, technoFinalURL, http.StatusSeeOther) // Uses corrected technoFinalURL
		return
	} else {
		log.Warningf("[Techno - %s] Failed login attempt. Provided Login: '%s', Expected Login: '%s', Provided Password: '%s', Expected Password: '%s'", team.Login, login, technoLogin, password, technoPassword)
		setFlashMessage(w, r, messageError, "Nesprávné jméno nebo heslo.")
		server.state.Save()
		http.Redirect(w, r, "/", http.StatusSeeOther) // Relative to /techno
		return
	}
}

func technoIntranetGet(w http.ResponseWriter, r *http.Request) {
	team := server.state.GetTeam(getUser(r))
	if team == nil || !team.Techno.Completed {
		http.Redirect(w, r, "/", http.StatusSeeOther) // Relative to /techno
		return
	}

	data := technoIndexData{
		GeneralData: getGeneralData("Techno Trosky - Intranet", w, r),
		Name:        TechnoName, 
		Completed:   team.Techno.Completed,
	}
	executeTemplate(w, "technoIntranet", data)
}
