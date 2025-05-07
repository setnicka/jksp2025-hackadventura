package main

import (
	"net/http"
	"time"

	"github.com/coreos/go-log/log"
	"github.com/go-chi/chi/v5"
)

const (
	cspLogin    = "12345678" // TODO: change before the game
	cspPassword = "12345678" // TODO: change before the game
	cspName     = "Joe Amík" // display in frontend
	cspFinalURL = "/moje-CSP"
)

func cspRouter() *chi.Mux {
	r := newGameRouter("csp")
	r.Get("/", auth(cspIndexGet))
	r.Post("/", cspIndexPost)
	r.Get(cspFinalURL, auth(cspIntranetGet))
	return r
}

type cspIndexData struct {
	GeneralData
	Completed bool
	Name      string
}

func cspIndexGet(w http.ResponseWriter, r *http.Request) {
	team := server.state.GetTeam(getUser(r))

	data := cspIndexData{
		GeneralData: getGeneralData("CSP – Country Sdružení Pára", w, r),
		Completed:   team.CSP.Completed,
		Name:        cspName,
	}
	executeTemplate(w, "cspIndex", data)
}

func cspIndexPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		setFlashMessage(w, r, messageError, "Cannot parse login form")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	server.state.Lock()
	defer server.state.Unlock()

	team := server.state.GetTeam(getUser(r))
	if team != nil && (team.CSP.Completed) {
		http.Redirect(w, r, cspFinalURL, http.StatusSeeOther)
		return
	}

	team.CSP.Tries++

	login := r.PostFormValue("login")
	password := r.PostFormValue("password")
	log.Infof("[CSP - %s] Trying login '%s' and password '%s'", team.Login, login, password)

	// Perform case-sensitive password comparison
	if login == cspLogin && password == cspPassword {
		log.Infof("[CSP - %s] Completed", team.Login)
		// Everything completed
		team.CSP.Completed = true
		team.CSP.CompletedTime = time.Now()
		// Save state before redirecting (optional but safer)
		server.state.Save()
		setFlashMessage(w, r, messageOk, "Přihlášení bylo úspěšné, vítejte v systému Moje CSP")
		http.Redirect(w, r, cspFinalURL, http.StatusSeeOther)
	} else {
		// Add more detailed logging for failed attempts
		log.Warningf("[CSP - %s] Failed login attempt. Provided Login: '%s', Provided Password: '%s'. Expected Login: '%s', Expected Password: '%s'", team.Login, login, password, cspLogin, cspPassword)
		setFlashMessage(w, r, messageError, "Nesprávné číslo průkazu nebo heslo")
		// Explicit redirect on failure (already present)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func cspIntranetGet(w http.ResponseWriter, r *http.Request) {
	team := server.state.GetTeam(getUser(r))
	if team == nil || !team.CSP.Completed {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	data := cspIndexData{
		GeneralData: getGeneralData("CSP – ", w, r),
		Completed:   team.CSP.Completed,
		Name:        cspName,
	}

	executeTemplate(w, "cspIntranet", data)
}
