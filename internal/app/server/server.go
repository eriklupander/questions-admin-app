package server

import (
	"github.com/a-h/templ"
	"github.com/eriklupander/templ-demo/internal/app"
	"github.com/eriklupander/templ-demo/internal/app/store"
	"github.com/eriklupander/templ-demo/internal/app/views"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golangcollege/sessions"
	"net/http"
	"strconv"
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

func answerQuestionHandler(session *sessions.Session, db *store.InMem) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
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
	}
}

func deleteQuestionHandler(session *sessions.Session, db *store.InMem) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
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

	}
}

func answerQuestionPage(session *sessions.Session, db *store.InMem) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
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
	}
}

func myQuestionsHandler(session *sessions.Session, db *store.InMem) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		email := session.GetString(r, "email")
		if email == "" {
			http.Error(w, "not logged in", 401)
			return
		}
		questions := db.AllForAuthor(email)
		session.Put(r, "view", "mine")
		templ.Handler(views.Questions(questions)).ServeHTTP(w, r)
	}
}

func allQuestionsHandler(session *sessions.Session, db *store.InMem) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		email := session.GetString(r, "email")
		if email == "" {
			http.Error(w, "not logged in", 401)
			return
		}
		questions := db.All()
		session.Put(r, "view", "all")
		templ.Handler(views.Questions(questions)).ServeHTTP(w, r)
	}
}

func countOwnHandler(session *sessions.Session, db *store.InMem) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		email := session.GetString(r, "email")
		all := len(db.AllForAuthorInStatus(email, app.StatusOpen))
		_, _ = w.Write([]byte(" (" + strconv.Itoa(all) + ")"))
	}
}

func countAllHandler(db *store.InMem) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		all := len(db.AllInStatus(app.StatusOpen))
		_, _ = w.Write([]byte(" (" + strconv.Itoa(all) + ")"))
	}
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
