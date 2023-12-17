package store

import (
	"fmt"
	"github.com/brianvoe/gofakeit"
	"github.com/eriklupander/templ-demo/internal/app"
	"github.com/google/uuid"
	"log/slog"
	"math/rand"
	"slices"
	"sort"
	"time"
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
		return out[i].CreatedAt.Before(out[j].CreatedAt)
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
		return out[i].CreatedAt.Before(out[j].CreatedAt)
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
		return out[i].CreatedAt.Before(out[j].CreatedAt)
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
		return out[i].CreatedAt.Before(out[j].CreatedAt)
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
		return out[i].CreatedAt.Before(out[j].CreatedAt)
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

func (i *InMem) FakeData() {
	talk1 := app.Talk{
		ID:      "1",
		Title:   "Spring Boot 3.2",
		Authors: []string{"magnus.larsson@callistaenterprise.se"},
	}
	talk2 := app.Talk{
		ID:      "2",
		Title:   "Go with HTMX",
		Authors: []string{"erik.lupander@callistaenterprise.se"},
	}
	talk3 := app.Talk{
		ID:      "3",
		Title:   "AI - Skynet is here",
		Authors: []string{"bjorn.genfors@callistaenterprise.se"},
	}
	talks := []app.Talk{talk1, talk2, talk3}
	gofakeit.Seed(time.Now().UnixMilli())

	for j := 0; j < 15; j++ {
		rnd := rand.Intn(len(talks))
		q := app.Question{
			ID:        uuid.NewString(),
			Talk:      talks[rnd],
			From:      gofakeit.Name(),
			Text:      gofakeit.Question(),
			CreatedAt: time.Now(),
			Status:    app.StatusOpen,
		}
		i.db = append(i.db, q)
	}

	//go func() {
	//	for {
	//		time.Sleep(time.Second * 15)
	//		rnd := rand.Intn(len(talks))
	//		q := app.Question{
	//			ID:        uuid.NewString(),
	//			Talk:      talks[rnd],
	//			From:      gofakeit.Name(),
	//			Text:      gofakeit.Question(),
	//			CreatedAt: time.Now(),
	//			Status:    app.StatusOpen,
	//		}
	//		i.db = append(i.db, q)
	//	}
	//}()
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
	q.Answer = &app.Answer{AnsweredBy: answeredBy, Text: text, AnsweredAt: time.Now()}
	q.Status = app.StatusAnswered

	// Delete the old one and add the new
	i.Delete(id)
	i.Create(q)
	return nil
}
