package server

import (
	"bytes"
	"github.com/eriklupander/templ-demo/internal/app"
	"github.com/eriklupander/templ-demo/internal/app/store"
	"github.com/go-chi/chi/v5"
	"github.com/golangcollege/sessions"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestAnswerQuestionHandler(t *testing.T) {

	db := store.NewInMem()

	tcases := []struct {
		name         string
		id           string
		question     *app.Question
		loggedIn     bool
		expectStatus int
	}{
		{"Happy path", "123", &app.Question{ID: "123", Talk: app.Talk{ID: "def"}, Answers: make([]app.Answer, 0)}, true, http.StatusOK},
		{"Unknown ID", "456", &app.Question{ID: "123", Talk: app.Talk{ID: "def"}, Answers: make([]app.Answer, 0)}, true, http.StatusInternalServerError},
		{"Not logged in", "123", &app.Question{ID: "123", Talk: app.Talk{ID: "def"}, Answers: make([]app.Answer, 0)}, false, http.StatusUnauthorized},
	}

	for _, tc := range tcases {
		t.Run(tc.name, func(t *testing.T) {
			session := sessions.New([]byte(`unit-testing`))
			if tc.question != nil {
				db.Create(*tc.question)
			}

			// Setup router for test
			r := chi.NewRouter()
			r.Use(session.Enable)
			if tc.loggedIn {
				// Use a test-only middleware to seed the logged-in user.
				r.Use(func(h http.Handler) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						session.Put(r, "email", "john.doe@doemain.com")
						h.ServeHTTP(w, r)
					})
				})
			}
			r.Post("/answerquestion", answerQuestionHandler(session, db))

			formData := &url.Values{}
			formData.Set("id", tc.id)
			formData.Set("answertext", "I am a walrus")

			buf := bytes.NewBuffer([]byte(formData.Encode()))
			req := httptest.NewRequest("POST", "/answerquestion", buf)
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, tc.expectStatus, w.Result().StatusCode)
		})
	}

}
