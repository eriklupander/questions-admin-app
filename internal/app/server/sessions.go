package server

import (
	"encoding/json"
	"github.com/golangcollege/sessions"
	gsessions "github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	"log/slog"
	"net/http"
	"os"
	"time"
)

func SetupSessions() (*sessions.Session, error) {

	// User session mgmt, used once gothic has performed the social login.
	var session *sessions.Session
	var secret = []byte("u46IpCV9y5Vlur8YvODJEhgOY8m9JVE4")
	session = sessions.New(secret)
	session.Lifetime = 3 * time.Hour

	// Gothic session mgmt for social login
	key := "Secret-session-key" // Replace with your SESSION_SECRET or similar
	maxAge := 86400 * 30        // 30 days
	isProd := false             // Set to true when serving over https

	cookieStore := gsessions.NewCookieStore([]byte(key))
	cookieStore.MaxAge(maxAge)

	cookieStore.Options.Domain = "localhost"
	cookieStore.Options.Path = ""
	cookieStore.Options.HttpOnly = false // HttpOnly should always be enabled
	cookieStore.Options.Secure = isProd

	gothic.Store = cookieStore

	// Read Google auth credentials from .credentials file.
	clientID, clientSecret := readCredentials()

	goth.UseProviders(
		google.New(clientID, clientSecret, "http://localhost:3000/auth/google/callback", "email", "profile"),
	)

	return session, nil
}

func logoutHandler(session *sessions.Session) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		session.Destroy(r)
		gothic.Logout(w, r)
		w.Header().Set("HX-Redirect", "/") // Use the special HTMX redirect header to trigger a full page reload.
	}
}

func readCredentials() (string, string) {
	credzData, err := os.ReadFile(".credentials")
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	credz := make(map[string]interface{})
	if err := json.Unmarshal(credzData, &credz); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	clientID := credz["web"].(map[string]interface{})["client_id"].(string)
	clientSecret := credz["web"].(map[string]interface{})["client_secret"].(string)
	return clientID, clientSecret
}
