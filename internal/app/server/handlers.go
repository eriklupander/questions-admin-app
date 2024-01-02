package server

import (
	"github.com/a-h/templ"
	"github.com/eriklupander/templ-demo/internal/app"
	"github.com/eriklupander/templ-demo/internal/app/store"
	"github.com/eriklupander/templ-demo/internal/app/views"
	"github.com/golangcollege/sessions"
	"net/http"
	"strconv"
)

func answerQuestionHandler(session *sessions.Session, db *store.InMem) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		email := session.GetString(r, "email")
		if email == "" {
			http.Error(w, "not logged in", 401)
			return
		}

		questionID := r.FormValue("id")
		answerText := r.FormValue("answertext")
		err := db.SaveAnswer(questionID, answerText, email)

		if err != nil {
			http.Error(w, "error saving answer", http.StatusInternalServerError)
			return
		}

		indexPage(session, db)(w, r)
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
		questionID := r.FormValue("id")
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
