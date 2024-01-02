package server

import (
	"github.com/a-h/templ"
	"github.com/eriklupander/templ-demo/internal/app/store"
	"github.com/eriklupander/templ-demo/internal/app/views"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golangcollege/sessions"
	"net/http"
)

func StartServer(session *sessions.Session, db *store.InMem) {
	// Set-up chi router with middleware
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(session.Enable)

	// Page specific handlers
	r.Get("/", indexPage(session, db))
	r.Get("/login", templ.Handler(views.Login()).ServeHTTP)
	r.Get("/answer", answerQuestionPage(session, db))

	// Social login handlers
	r.Get("/auth", authStartHandler())
	r.Get("/auth/{provider}/callback", authCallbackHandler(session))
	r.Get("/logout", logoutHandler(session))

	// API handlers
	r.Get("/countall", countAllHandler(db))
	r.Get("/countmine", countOwnHandler(session, db))

	r.Get("/all", allQuestionsHandler(session, db))
	r.Get("/mine", myQuestionsHandler(session, db))

	r.Post("/answerquestion", answerQuestionHandler(session, db))
	r.Delete("/delete", deleteQuestionHandler(session, db))

	// Start plain HTTP listener
	_ = http.ListenAndServe(":3000", r)
}

func indexPage(session *sessions.Session, db *store.InMem) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		email := session.GetString(r, "email")
		if email != "" {
			if session.GetString(r, "view") == "mine" {
				templ.Handler(views.Index(email, db.AllForAuthor(email))).ServeHTTP(w, r)
			} else {
				templ.Handler(views.Index(email, db.All())).ServeHTTP(w, r)
			}
			return
		}
		templ.Handler(views.Index("", nil)).ServeHTTP(w, r)
	}
}
