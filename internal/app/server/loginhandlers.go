package server

import (
	"github.com/golangcollege/sessions"
	"github.com/markbates/goth/gothic"
	"net/http"
)

func authCallbackHandler(session *sessions.Session) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		user, err := gothic.CompleteUserAuth(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		session.Put(r, "email", user.Email)
		session.Put(r, "view", "all")
		http.Redirect(w, r, "/", 302)
	}
}

func authStartHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		gothic.BeginAuthHandler(w, r)
	}
}
