package main

import (
	"github.com/eriklupander/templ-demo/internal/app/server"
	"github.com/eriklupander/templ-demo/internal/app/store"
	"log/slog"
	"os"
)

func main() {

	session, err := server.SetupSessions()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	// In-memory DB for questions and answers.
	db := store.NewInMem()
	db.SeedWithFakeData()

	// Setup and start HTTP server
	server.StartServer(session, db)
}
