package api

import (
	"net/http"
	"time"

	"scheduler/pkg/db"
)

type taskAPI struct {
	store *db.Store
}

func Register(mux *http.ServeMux, store *db.Store, password string) error {
	auth, err := newAuth(password)
	if err != nil {
		return err
	}

	api := &taskAPI{store: store}
	mux.HandleFunc("GET /api/nextdate", nextDateHandler)
	mux.HandleFunc("POST /api/signin", auth.signin)
	mux.HandleFunc("GET /api/tasks", auth.require(api.listTasks))
	mux.HandleFunc("GET /api/task", auth.require(api.getTask))
	mux.HandleFunc("POST /api/task", auth.require(api.addTask))
	mux.HandleFunc("PUT /api/task", auth.require(api.editTask))
	mux.HandleFunc("DELETE /api/task", auth.require(api.deleteTask))
	mux.HandleFunc("POST /api/task/done", auth.require(api.completeTask))
	return nil
}

func nextDateHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	now := time.Now()
	if raw := query.Get("now"); raw != "" {
		parsed, err := time.Parse(dateLayout, raw)
		if err != nil {
			http.Error(w, "Параметр now должен содержать дату в формате ГГГГММДД", http.StatusBadRequest)
			return
		}
		now = parsed
	}

	date, err := NextDate(now, query.Get("date"), query.Get("repeat"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte(date))
}
