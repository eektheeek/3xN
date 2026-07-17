package main

import (
	"log"
	"net/http"
	"os"

	"github.com/eektheeek/dead-lift-project/diary-api/internal/db"
	"github.com/eektheeek/dead-lift-project/diary-api/internal/handlers"
	"github.com/eektheeek/dead-lift-project/diary-api/internal/repository"
)

func main() {
	dbPath := envOr("DIARY_DB", "data/diary.db")
	migrationsDir := envOr("DIARY_MIGRATIONS", "migrations")
	addr := envOr("DIARY_ADDR", ":8080")

	sqlDB, err := db.Open(dbPath, migrationsDir)
	if err != nil {
		log.Fatalf("db open: %v", err)
	}
	defer sqlDB.Close()

	repo := repository.New(sqlDB)
	api := handlers.NewAPI(repo)
	mux := handlers.NewMux(api)

	log.Printf("listening on %s (db=%s)", addr, dbPath)
	if err := http.ListenAndServe(addr, handlers.CORS(mux)); err != nil {
		log.Fatalf("listen: %v", err)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
