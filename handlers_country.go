package main

import (
	"net/http"
	"time"

	"github.com/coreos/go-log/log"
	"github.com/go-chi/chi/v5"
)

const (
	cspLogin    = "csp_mamik@seznam.cz" // TODO: change before the game
	cspName     = "Joe Amík" // display in frontend
	cspFinalURL = "/moje-CSP"
	cspFailURL = "/csp-fail"
)

func cspRouter() *chi.Mux {
	r := newGameRouter("csp")
	r.Get("/", auth(cspIndexGet))
	r.Post("/", cspIndexPost)
	r.Get("/csp-fail", auth(cspFailGet))
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

	var was time.Time
	lastTrue := false

	if login == "admin" && password == "admin" {
		log.Infof("[CSP - %s] Matush login haha", team.Login)
		http.Redirect(w, r, cspFailURL, http.StatusSeeOther)
		return
	}

	for _, p := range server.state.GetCSPPasswords() {
		if p.Password == password && cspLogin == login {
			was = p.From
			lastTrue = true
		} else {
			lastTrue = false
		}
	}

	if lastTrue {
		log.Infof("[CSP - %s] Completed", team.Login)
		team.CSP.Completed = true
		team.CSP.CompletedTime = time.Now()
		http.Redirect(w, r, cspFinalURL, http.StatusSeeOther)
		return
	} else if !was.IsZero() {
		d := time.Since(was)
		if d.Seconds() < 5 {
			setFlashMessage(w, r, messageWarn, "Tohle heslo jste právě přestali používat, vždyť jste ho teď změnil! Kvůli bezpečnosti vás nepustíme!")
		} else if d.Seconds() > 60 {
			setFlashMessage(w, r, messageWarn, "Tohle je staré heslo, to jste používali před %d minutami! Kvůli bezpečnosti vás nepustíme!", int(d.Minutes()))
		} else {
			setFlashMessage(w, r, messageWarn, "Tohle je staré heslo, to jste používali před %d sekundami! Kvůli bezpečnosti vás nepustíme!", int(d.Seconds()))
		}
	} else {
		setFlashMessage(w, r, messageWarn, "Cože?!? To ses opil kofolou?")
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
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
func cspFailGet(w http.ResponseWriter, r *http.Request) {
	team := server.state.GetTeam(getUser(r))

	data := cspIndexData{
		GeneralData: getGeneralData("CSP – ", w, r),
		Completed:   team.CSP.Completed,
		Name:        cspName,
	}

	executeTemplate(w, "cspfail", data)
}
