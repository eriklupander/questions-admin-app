package store

import (
	"fmt"
	"log/slog"
	"math/rand"
	"slices"
	"sort"
	"time"

	"github.com/brianvoe/gofakeit"
	"github.com/eriklupander/templ-demo/internal/app"
	"github.com/google/uuid"
)

type InMem struct {
	db []app.Question
}

func NewInMem() *InMem {
	db := make([]app.Question, 0)
	return &InMem{db: db}
}

func (i *InMem) Create(q app.Question) {
	i.db = append(i.db, q)
}
func (i *InMem) All() []app.Question {
	out := i.db
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.After(out[j].CreatedAt)
		//return out[i].CreatedAt.After(out[j].CreatedAt) && out[i].Status == app.StatusOpen
	})
	return out
}
func (i *InMem) AllForTalk(talkID string) []app.Question {
	out := make([]app.Question, 0)
	for _, q := range i.db {
		if q.Talk.ID == talkID {
			out = append(out, q)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out
}
func (i *InMem) AllForTalkInStatus(talkID string, status app.Status) []app.Question {
	out := make([]app.Question, 0)
	for _, q := range i.db {
		if q.Talk.ID == talkID && q.Status == status {
			out = append(out, q)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out
}

func (i *InMem) AllInStatus(status app.Status) []app.Question {
	out := make([]app.Question, 0)
	for _, q := range i.db {
		if q.Status == status {
			out = append(out, q)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		//return (out[j].Status == app.StatusOpen && out[i].Status != app.StatusOpen) && out[i].CreatedAt.After(out[j].CreatedAt)
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out
}

func (i *InMem) AllForAuthor(email string) []app.Question {
	out := make([]app.Question, 0)
	for _, q := range i.db {
		if slices.Contains(q.Talk.Authors, email) {
			out = append(out, q)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out
}

func (i *InMem) AllForAuthorInStatus(email string, status app.Status) []app.Question {
	out := make([]app.Question, 0)
	for _, q := range i.db {
		if slices.Contains(q.Talk.Authors, email) && q.Status == status {
			out = append(out, q)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out
}

func (i *InMem) Delete(id string) {
	slog.Info("delete question by id", slog.String("id", id))
	for j := range i.db {
		if i.db[j].ID == id {
			i.db = slices.Delete(i.db, j, j+1)
			break
		}
	}
}

func (i *InMem) SeedWithFakeData() {
	talk1 := app.Talk{
		ID:      "1",
		Title:   "Spring Boot 3.2",
		Authors: []string{"hanna.hansson@callistaenterprise.se"},
	}
	talk2 := app.Talk{
		ID:      "2",
		Title:   "Go with HTMX",
		Authors: []string{"erik.lupander@callistaenterprise.se"},
	}
	talk3 := app.Talk{
		ID:      "3",
		Title:   "AI - Skynet is here",
		Authors: []string{"kalle.karlsson@callistaenterprise.se"},
	}
	talks := []app.Talk{talk1, talk2, talk3}
	gofakeit.Seed(time.Now().UnixMilli())

	for j := 0; j < 15; j++ {
		rnd := rand.Intn(len(talks))
		negRand := -rand.Intn(120)
		q := app.Question{
			ID:        uuid.NewString(),
			Talk:      talks[rnd],
			From:      gofakeit.Name(),
			Text:      gofakeit.Question(),
			CreatedAt: time.Now().Add(time.Minute * time.Duration(negRand)),
			Status:    app.StatusOpen,
		}
		i.db = append(i.db, q)
	}

}

func (i *InMem) Get(id string) (app.Question, error) {
	for _, q := range i.db {
		if q.ID == id {
			return q, nil
		}
	}
	return app.Question{}, fmt.Errorf("question identified by %v not found", id)
}

func (i *InMem) SaveAnswer(id string, text string, answeredBy string) error {
	q, err := i.Get(id)
	if err != nil {
		return err
	}
	q.Answers = append(q.Answers, app.Answer{ID: uuid.NewString(), AnsweredBy: answeredBy, Text: text, AnsweredAt: time.Now()})
	q.Status = app.StatusAnswered

	// Delete the old one and add the new
	i.Delete(id)
	i.Create(q)
	return nil
}

func (i *InMem) Emit(qChan chan app.Question) {

	go func() {
		for {
			time.Sleep(time.Second * 3)
			q := app.Question{
				ID:        uuid.NewString(),
				Talk:      app.Talk{ID: uuid.NewString(), Title: gofakeit.HipsterSentence(3)},
				From:      gofakeit.Name(),
				Text:      gofakeit.Question(),
				CreatedAt: time.Now(),
				Status:    app.StatusOpen,
			}
			i.db = append(i.db, q)
			qChan <- q
		}
	}()
}
