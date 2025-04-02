package main

import (
	"encoding/gob"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
)

func auth(handle http.HandlerFunc, renewAuth ...bool) http.HandlerFunc {
	renew := true
	if len(renewAuth) > 0 {
		renew = renewAuth[0]
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if checkSession(w, r, renew) {
			if server.state.GetTeam(getUser(r)) != nil {
				handle(w, r)
				return
			}
		}
		http.Redirect(w, r, "/start-hry", http.StatusTemporaryRedirect)
	}
}

func getUser(r *http.Request) string {
	login, _ := getSession(r).Values["user"].(string)
	return login
}

func getSession(r *http.Request) *sessions.Session {
	session, err := server.sessionStore.Get(r, sessionCookieName)
	if err != nil {
		slog.Error("cannot get session", "name", sessionCookieName, "err", err)
		return nil
	}
	return session
}

func checkSession(w http.ResponseWriter, r *http.Request, renew bool) bool {
	session := getSession(r)
	if session == nil {
		return false
	}
	// Check if user is authenticated
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		return false
	}

	if renew {
		session.Save(r, w)
	}
	return true
}

////////////////////////////////////////////////////////////////////////////////

const flashSessionStoreName = "flash-session"

type flashMessageType string // mapping to bootstrap CSS classes

const (
	messageOk    flashMessageType = "success"
	messageError flashMessageType = "danger"
	messageWarn  flashMessageType = "warning"
)

// FlashMessage holds type and content of flash message displayed to the user
type FlashMessage struct {
	Type    flashMessageType
	Message string
}

func setFlashMessage(w http.ResponseWriter, r *http.Request, messageType flashMessageType, message string, args ...interface{}) {
	// Register the struct so encoding/gob knows about it
	gob.Register(FlashMessage{})

	session, err := server.sessionStore.Get(r, flashSessionStoreName)
	if err != nil {
		return
	}
	session.AddFlash(FlashMessage{Type: messageType, Message: fmt.Sprintf(message, args...)})
	err = session.Save(r, w)
	if err != nil {
		slog.Error("cannot save flash message", "err", err)
	}
}

func getFlashMessages(w http.ResponseWriter, r *http.Request) []FlashMessage {
	// 1. Get session
	session, err := server.sessionStore.Get(r, flashSessionStoreName)
	if err != nil {
		return nil
	}

	// 2. Get flash messages
	parsedFlashes := []FlashMessage{}
	if flashes := session.Flashes(); len(flashes) > 0 {
		for _, flash := range flashes {
			parsedFlashes = append(parsedFlashes, flash.(FlashMessage))
		}
	}

	// 3. Delete flash messages by saving session
	err = session.Save(r, w)
	if err != nil {
		slog.Error("problem during loading flash messages", "err", err)
	}

	return parsedFlashes
}

////////////////////////////////////////////////////////////////////////////////

type noListFile struct {
	http.File
}

func (f noListFile) Readdir(_ int) ([]os.FileInfo, error) {
	return nil, nil
}

// NoListFileSystem is like http.FileSystem but it does not provide directory
// listing
type NoListFileSystem struct {
	base http.FileSystem
}

// Open given file, for directories it returns noListFile instance which does
// not provide directory listing
func (fs NoListFileSystem) Open(name string) (http.File, error) {
	f, err := fs.base.Open(name)
	if err != nil {
		return nil, err
	}
	return noListFile{f}, nil
}
