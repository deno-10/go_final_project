package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"scheduler/pkg/api"
	"scheduler/pkg/db"
)

func Run() error {
	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = "7540"
	}
	if value, err := strconv.Atoi(port); err != nil || value < 1 || value > 65535 {
		return fmt.Errorf("некорректный порт TODO_PORT: %q", port)
	}

	dbFile := os.Getenv("TODO_DBFILE")
	if dbFile == "" {
		dbFile = "scheduler.db"
	}
	store, err := db.Open(dbFile)
	if err != nil {
		return fmt.Errorf("открытие базы данных: %w", err)
	}
	defer store.Close()

	mux := http.NewServeMux()
	if err := api.Register(mux, store, os.Getenv("TODO_PASSWORD")); err != nil {
		return err
	}
	mux.Handle("GET /", http.FileServer(http.Dir("web")))

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	log.Printf("Планировщик доступен по адресу http://localhost:%s/", port)
	return srv.ListenAndServe()
}
