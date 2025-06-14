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
	listen     = flag.String("listen", ":8080", "Listen address")
	baseDomain = flag.String("domain", "localhost:8080", "Base domain")
)

// Global singleton
var server *Server

////////////////////////////////////////////////////////////////////////////////

type targetS struct {
	Code     string
	Name     string
	URL      string
	Router   func() *chi.Mux
	GetState func(t *state.Team) state.Result
}

var targets []targetS

////////////////////////////////////////////////////////////////////////////////

func (sub subdomains) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	domainParts := strings.Split(r.Host, ".")

	if mux := sub[domainParts[0]]; mux != nil {
		// Let the appropriate mux serve the request
		mux.ServeHTTP(w, r)
	} else {
		executeTemplate(w, "noSubdomainIndex", targets)
	}
}

func main() {
	flag.Parse()

	targets = []targetS{
		{Code: "metal", Name: "Metal", URL: "metal." + *baseDomain,
			Router:   metalRouter,
			GetState: func(t *state.Team) state.Result { return t.Metal.Result }},
		{Code: "hotel", Name: "Hotel", URL: "hotel." + *baseDomain,
			Router:   hotelRouter,
			GetState: func(t *state.Team) state.Result { return t.Hotel.Result }},
		{Code: "kpop", Name: "K-Pop", URL: "kpop." + *baseDomain,
			Router:   kpopRouter,
			GetState: func(t *state.Team) state.Result { return t.Kpop.Result }},
		{Code: "satna", Name: "Satna", URL: "satna." + *baseDomain,
			Router:   satnaRouter,
			GetState: func(t *state.Team) state.Result { return t.Satna.Result }},
		{Code: "csp", Name: "Country", URL: "csp." + *baseDomain,
			Router:   cspRouter,
			GetState: func(t *state.Team) state.Result { return t.CSP.Result }},
		{Code: "techno", Name: "Techno", URL: "techno." + *baseDomain,
			Router:   technoRouter,
			GetState: func(t *state.Team) state.Result { return t.Techno.Result }},
		{Code: "klasicka", Name: "Klasicka", URL: "klasicka." + *baseDomain,
			Router:   klasickaRouter,
			GetState: func(t *state.Team) state.Result { return t.Klasicka.Result }},
	}

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
	slog.Info("starting server", "listen", *listen, "baseDomain", *baseDomain)
	if err := http.ListenAndServe(*listen, r); err != nil {
		slog.Error("SERVER ERROR", "err", err)
	}
}
