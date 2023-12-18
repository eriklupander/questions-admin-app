package main

import (
	"encoding/json"
	"github.com/a-h/templ"
	"github.com/eriklupander/templ-demo/internal/app"
	"github.com/eriklupander/templ-demo/internal/app/store"
	"github.com/eriklupander/templ-demo/internal/app/views"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golangcollege/sessions"
	gsessions "github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {

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

	// In-memory DB for questions and answers.
	db := store.NewInMem()
	db.FakeData()

	// Set-up chi router with middleware
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(session.Enable)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		email := session.GetString(r, "email")
		if email != "" {
			if session.GetString(r, "view") == "mine" {
				templ.Handler(views.Index(email, db.AllForAuthor(email))).ServeHTTP(w, r)
			} else {
				templ.Handler(views.Index(email, db.All())).ServeHTTP(w, r)
			}
			return
		}
		templ.Handler(views.Index(email, db.All())).ServeHTTP(w, r)
	})

	r.Get("/auth", func(w http.ResponseWriter, r *http.Request) {
		gothic.BeginAuthHandler(w, r)
	})
	r.Get("/logout", func(w http.ResponseWriter, r *http.Request) {
		session.Destroy(r)
		gothic.Logout(w, r)
		http.Redirect(w, r, "/", 302)
	})

	r.Get("/auth/{provider}/callback", func(w http.ResponseWriter, r *http.Request) {

		user, err := gothic.CompleteUserAuth(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		session.Put(r, "email", user.Email)
		session.Put(r, "view", "all")
		http.Redirect(w, r, "/", 302)
	})

	r.Get("/login", templ.Handler(views.Login()).ServeHTTP)
	r.Get("/countall", func(w http.ResponseWriter, r *http.Request) {
		all := len(db.AllInStatus(app.StatusOpen))
		_, _ = w.Write([]byte(" (" + strconv.Itoa(all) + ")"))
	})
	r.Get("/countmine", func(w http.ResponseWriter, r *http.Request) {
		email := session.GetString(r, "email")
		all := len(db.AllForAuthorInStatus(email, app.StatusOpen))
		_, _ = w.Write([]byte(" (" + strconv.Itoa(all) + ")"))
	})
	r.Get("/all", func(w http.ResponseWriter, r *http.Request) {
		email := session.GetString(r, "email")
		if email == "" {
			http.Error(w, "not logged in", 401)
			return
		}
		questions := db.All()
		session.Put(r, "view", "all")
		templ.Handler(views.Questions(questions)).ServeHTTP(w, r)
	})
	r.Get("/mine", func(w http.ResponseWriter, r *http.Request) {
		email := session.GetString(r, "email")
		if email == "" {
			http.Error(w, "not logged in", 401)
			return
		}
		questions := db.AllForAuthor(email)
		session.Put(r, "view", "mine")
		templ.Handler(views.Questions(questions)).ServeHTTP(w, r)
	})

	// Buttons on the card

	r.Get("/answer", func(w http.ResponseWriter, r *http.Request) {
		email := session.GetString(r, "email")
		if email == "" {
			http.Error(w, "not logged in", 401)
			return
		}

		questionID := r.URL.Query().Get("id")
		question, err := db.Get(questionID)
		if err != nil {
			http.Error(w, "question not found", 404)
			return
		}
		templ.Handler(views.AnswerQuestion(email, question)).ServeHTTP(w, r)
	})

	r.Post("/answerquestion", func(w http.ResponseWriter, r *http.Request) {
		email := session.GetString(r, "email")
		if email == "" {
			http.Error(w, "not logged in", 401)
			return
		}

		questionID := r.URL.Query().Get("id")
		answerText := r.FormValue("answertext")
		err := db.SaveAnswer(questionID, answerText, email)

		if err != nil {
			http.Error(w, "error saving answer", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/", 302)
	})

	r.Delete("/delete", func(w http.ResponseWriter, r *http.Request) {
		email := session.GetString(r, "email")
		if email == "" {
			http.Error(w, "not logged in", 401)
			return
		}

		questionID := r.URL.Query().Get("id")
		db.Delete(questionID)
		if session.GetString(r, "view") == "mine" {
			templ.Handler(views.Questions(db.AllForAuthor(email))).ServeHTTP(w, r)
		} else {
			templ.Handler(views.Questions(db.All())).ServeHTTP(w, r)
		}

	})
	_ = http.ListenAndServe(":3000", r)
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
