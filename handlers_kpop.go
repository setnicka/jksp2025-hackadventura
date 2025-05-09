package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/coreos/go-log/log"
	"github.com/go-chi/chi/v5"
)

type moriaQuestion struct {
	Question string
	Answer   string
}

var moriaQuestions = []moriaQuestion{
	{"Začneme jednoduchými. Kolikrát mrknul zpěvák Rap Beast ze skupiny BFS během koncertu Find yourself v Tokyu v roce 2014?", "217"},
	{"Jaký je nejoblíbenější nápoj zpěváka Gimin?", "vodka s ledem"},
	{"Jaká byla oblíbená náhražka za K-Pop v Sovětském svazu?", "proleto-pop"},
	{"Kolik aminokyselin obsahuje DNA zpěváka J-Believe ze skupiny DFS?", "0"},
	{"Kolik různých beatů má písnička TNT od skupiny DFS?", "117"},
	{"Jaký zpěvák ze skupiny BFS je nejoblíbenější u ženského pohlaví?", "raptor"},
}

var moriaFinalURL = "/" + url.PathEscape("real-kpop-fan")

func kpopRouter() *chi.Mux {
	r := newGameRouter("kpop")
	r.Get("/", auth(moriaIndexGet))
	r.Post("/", auth(moriaIndexPost))
	r.Get("/real-kpop-fan", auth(moriaInternalGet))
	return r
}

type moriaIndexData struct {
	GeneralData
	Completed bool
	Image     string
    Music     string
	Script    string
	Question  string
}

func moriaIndexGet(w http.ResponseWriter, r *http.Request) {
	team := server.state.GetTeam(getUser(r))
	if team != nil && (team.Kpop.Completed) {
		http.Redirect(w, r, moriaFinalURL, http.StatusSeeOther)
		return
	}
	if team.Kpop.RightAnswers >= len(moriaQuestions) {
		http.Error(w, "Chybí další otázka, běž za orgy", 500)
		return
	}

	
	data := moriaIndexData{
		GeneralData: getGeneralData("K-Pop quiz", w, r),
		Image:       fmt.Sprintf("rm_disturbed_nino.jpg"),
		Music:       "BTS-Butter-YouTube.mp3",
		Script:      "script.js",
		Question:    moriaQuestions[team.Kpop.RightAnswers].Question,
	}

	getGeneralData("K-Pop quiz", w, r)
	executeTemplate(w, "kpopIndex", data)
}

func moriaIndexPost(w http.ResponseWriter, r *http.Request) {
	defer http.Redirect(w, r, "/", http.StatusSeeOther)
	if err := r.ParseForm(); err != nil {
		setFlashMessage(w, r, messageError, "Cannot parse form")
		return
	}

	server.state.Lock()
	defer server.state.Unlock()

	team := server.state.GetTeam(getUser(r))
	if team != nil && (team.Kpop.Completed) {
		http.Redirect(w, r, moriaFinalURL, http.StatusSeeOther)
		return
	}
	if team.Kpop.RightAnswers >= len(moriaQuestions) {
		http.Error(w, "Chybí další otázka, běž za orgy", 500)
		return
	}

	team.Kpop.Tries++
	defer server.state.Save()

	answer := r.PostFormValue("answer")
	log.Infof("[Kpop - %s] Trying answer '%s' to question %d", team.Login, answer, team.Kpop.RightAnswers)
	// normalization
	answer = strings.ToLower(strings.Replace(answer, ",", ".", 1))

	if answer == strings.ToLower(moriaQuestions[team.Kpop.RightAnswers].Answer) {
		log.Infof("[Kpop - %s] Right answer to question %d", team.Login, team.Kpop.RightAnswers)
		team.Kpop.RightAnswers++

		if team.Kpop.RightAnswers == len(moriaQuestions) {
			log.Infof("[Kpop - %s] Completed", team.Login)
			team.Kpop.Completed = true
			team.Kpop.CompletedTime = time.Now()
			http.Redirect(w, r, moriaFinalURL, http.StatusSeeOther)
		} else {
			setFlashMessage(w, r, messageOk, "Skvěle! Jen skutečný K-Pop fanoušek by věděl tuto odpověď.")
		}
	} else {
		setFlashMessage(w, r, messageError, "Děláš si *****? To si říkáš fanoušek? I moje babička toho ví víc o K-Popu než ty.")
	}
}

func moriaInternalGet(w http.ResponseWriter, r *http.Request) {
	team := server.state.GetTeam(getUser(r))
	if team == nil || !team.Kpop.Completed {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	data := moriaIndexData{
		GeneralData: getGeneralData("Real ones", w, r),
		Image:       fmt.Sprintf("rm_happy.jpg"),
		Script:      "script.js",
		Completed:   true,
	}

	executeTemplate(w, "kpopIndex", data)
}
