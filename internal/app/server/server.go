package server

import (
	"bytes"
	"fmt"
	"github.com/a-h/templ"
	"github.com/eriklupander/templ-demo/internal/app"
	"github.com/eriklupander/templ-demo/internal/app/store"
	"github.com/eriklupander/templ-demo/internal/app/views"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golangcollege/sessions"
	"log/slog"
	"net/http"
)

func StartServer(session *sessions.Session, db *store.InMem, qChan chan app.Question) {
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

	// SSE experiment
	r.Get("/stream", templ.Handler(views.Stream()).ServeHTTP)
	r.Get("/sse", func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "SSE is not supported", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

	OUTER:
		for {
			select {
			case q := <-qChan:
				buf := new(bytes.Buffer)

				buf.WriteString(fmt.Sprintf("event: %s\n", "question"))
				buf.WriteString("data: ")
				views.Card(q).Render(r.Context(), buf)
				buf.WriteString("\n\n")

				_, err := w.Write(buf.Bytes())
				if err != nil {
					slog.Error("SSE write failed", slog.Any("error", err))
				}
				flusher.Flush()
			case <-r.Context().Done():
				break OUTER
			}

		}
		slog.Info("SSE disconnect")
	})

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
