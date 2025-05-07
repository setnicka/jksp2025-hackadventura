package main

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

const (
	hotelLogin    = "AlexejIvanovic"
	hotelPassword = "deRatyzatoR1" // musí mít sudý počet znaků
)

func hotelRouter() *chi.Mux {
	r := newGameRouter("hotel")
	r.Get("/", auth(hotelIndexGet))
	r.Get("/interni", auth(hotelInternalGet))
	r.Post("/interni", auth(hotelInternalPost))
	return r
}

func b(fail bool) string {
	if fail {
		return "x"
	}
	return "."
}

func hotelIndexGet(w http.ResponseWriter, r *http.Request) {
	data := getGeneralData("El Hippo Grande", w, r)
	executeTemplate(w, "hotelIndex", data)
}

func hotelInternalGet(w http.ResponseWriter, r *http.Request) {
	team := server.state.GetTeam(getUser(r))

	if team != nil && team.Hotel.Completed {
		data := getGeneralData("HACKED - El Hippo Grande", w, r)
		executeTemplate(w, "hotelHacked", data)

	} else {
		data := getGeneralData("Pro zaměstnancov", w, r)
		executeTemplate(w, "hotelInternal", data)
	}
}

func hotelInternalPost(w http.ResponseWriter, r *http.Request) {
	defer http.Redirect(w, r, "/interni", http.StatusSeeOther)

	if err := r.ParseForm(); err != nil {
		setFlashMessage(w, r, messageError, "Cannot parse form")
		return
	}

	server.state.Lock()
	defer server.state.Unlock()

	team := server.state.GetTeam(getUser(r))
	if team != nil && (team.Hotel.Completed) {
		return
	}

	slog := slog.With("team", team.Login)

	team.Hotel.Tries++
	team.Hotel.LastTry = time.Now()
	defer server.state.Save()

	////////////////////////////////////////////////////////////////

	a1 := strings.TrimSpace(r.PostFormValue("a1"))   // první půlka hesla
	a2 := strings.TrimSpace(r.PostFormValue("a2"))   // login bez prvního písmenka
	a3 := strings.TrimSpace(r.PostFormValue("a3"))   // login
	a4 := strings.TrimSpace(r.PostFormValue("a4"))   // prázdné
	a5 := strings.TrimSpace(r.PostFormValue("a5"))   // 2021 (letopočet minulého léta)
	a6 := strings.TrimSpace(r.PostFormValue("a6"))   // druhý znak hesla
	a7 := strings.TrimSpace(r.PostFormValue("a7"))   // password
	a8 := strings.TrimSpace(r.PostFormValue("a8"))   // cucoriedka
	a9 := strings.TrimSpace(r.PostFormValue("a9"))   // login pozpátku
	a10 := strings.TrimSpace(r.PostFormValue("a10")) // každý druhý znak hesla

	pwHalf := hotelPassword[:len(hotelPassword)/2]
	loginWithoutFirst := hotelLogin[1:]
	loginReversed := ""
	for i := len(hotelLogin) - 1; i >= 0; i-- {
		loginReversed += string(hotelLogin[i])
	}
	pwEverySecond := ""
	for i := 0; i < len(hotelPassword); i += 2 {
		pwEverySecond += string(hotelPassword[i])
	}

	c1 := a1 != pwHalf
	c2 := a2 != loginWithoutFirst
	c3 := a3 != hotelLogin
	c4 := len(a4) != 0
	c5 := a5 != "2024"
	c6 := len(a6) < 1 || a6[0] != hotelPassword[1]
	c7 := a7 != hotelPassword
	c8 := strings.ToLower(a8) != "cucoriedka"
	c9 := a9 != loginReversed
	c10 := a10 != pwEverySecond

	tests := b(c1) + b(c2) + b(c3) + b(c4) + b(c5) + b(c6) + b(c7) + b(c8) + b(c9) + b(c10)
	slog.Info("[Hotel] unsuccessful try", "login", a3, "password", a7, "tests", tests)
	fail := c1 || c2 || c3 || c4 || c5 || c6 || c7 || c8 || c9 || c10

	if fail {
		testsFrontendOrder := b(c4) + b(c2) + b(c8) + b(c7) + b(c10) + b(c3) + b(c5) + b(c1) + b(c9) + b(c6)
		setFlashMessage(w, r, messageError, "Пассворда нет валидна "+testsFrontendOrder) // "Nesprávné údaje: ...x..."
		return
	}

	slog.Info("[Hotel] completed")
	// Everything completed
	team.Hotel.Completed = true
	team.Hotel.CompletedTime = time.Now()
}
