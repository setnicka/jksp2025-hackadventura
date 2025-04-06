// Server pro Hackadventuru na JKSP2025
package main

import (
	"flag"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/sessions"
	"github.com/lmittmann/tint"

	"jksp2025-hackadventura/state"
)

const (
	sessionCookieSecret = "vieShu6tis3gaiqu9faishibeipahz"
	sessionMaxAge       = 3600 * 24
	sessionCookieName   = "hack_systems"
	templatesDir        = "templates"
	staticDir           = "static"

	orgLogin    = "ksp"
	orgPassword = "ksp" // TODO: change before the game
)

// Server represents the whole system with its state
type Server struct {
	sessionStore sessions.Store
	templates    *template.Template
	state        *state.State
}

type subdomains map[string]*chi.Mux

var (
	listen = flag.String("listen", ":8080", "Listen address")
)

// Global singleton
var server *Server

////////////////////////////////////////////////////////////////////////////////

type targetS struct {
	Code   string
	Name   string
	URL    string
	Router func() *chi.Mux
}

// const baseDomain = "setnicka.cz:8080"
const baseDomain = "localhost:8080"

var targets = []targetS{
	{Code: "hotel", Name: "Hotel", URL: "hotel." + baseDomain, Router: hotelRouter},
	{Code: "satna", Name: "Satna", URL: "satna." + baseDomain, Router: satnaRouter},
}

////////////////////////////////////////////////////////////////////////////////

func (sub subdomains) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	domainParts := strings.Split(r.Host, ".")

	if mux := sub[domainParts[0]]; mux != nil {
		// Let the appropriate mux serve the request
		mux.ServeHTTP(w, r)
	} else {
		// Handle 404
		http.Error(w, "Subdomain not found", 404)
	}
}

func main() {
	cookieStore := sessions.NewCookieStore([]byte(sessionCookieSecret))
	cookieStore.MaxAge(sessionMaxAge)
	//cookieStore.Options.Domain = ".fuf.me"

	logger := slog.New(tint.NewHandler(os.Stderr, &tint.Options{}))
	slog.SetDefault(logger)

	server = &Server{
		sessionStore: cookieStore,
		state:        state.Init(),
	}

	server.start()
}

func (s *Server) start() {
	slog.Info("preparing server")

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	//r.Use(middleware.Logger) // used in subrouters independently
	r.Use(middleware.Recoverer)

	// Routers for subdomains:
	sub := subdomains{
		"org": orgRouter(),
	}
	for _, target := range targets {
		sub[target.Code] = target.Router()
	}

	if _, err := server.getTemplates(); err != nil {
		slog.Error("cannot load templates", "err", err)
		return
	}

	r.Mount("/", sub)
	slog.Info("starting server", "listen", *listen)
	if err := http.ListenAndServe(*listen, r); err != nil {
		slog.Error("SERVER ERROR", "err", err)
	}
}
