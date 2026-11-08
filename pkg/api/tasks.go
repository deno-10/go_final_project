package api

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"scheduler/pkg/db"
)

func checkID(w http.ResponseWriter, id string) bool {
	if id == "" {
		writeError(w, http.StatusBadRequest, "Не указан идентификатор задачи")
		return false
	}
	if num, err := strconv.ParseInt(id, 10, 64); err != nil || num < 1 {
		writeError(w, http.StatusBadRequest, "Идентификатор задачи должен быть положительным целым числом")
		return false
	}
	return true
}

func writeStoreError(w http.ResponseWriter, err error) {
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "Задача не найдена")
		return
	}
	log.Printf("Операция с задачей: %v", err)
	writeError(w, http.StatusInternalServerError, "Не удалось выполнить операцию с задачей")
}

func (api *taskAPI) listTasks(w http.ResponseWriter, r *http.Request) {
	const limit = 50
	tasks, err := api.store.Tasks(r.Context(), r.URL.Query().Get("search"), limit)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string][]db.Task{"tasks": tasks})
}

func (api *taskAPI) getTask(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if !checkID(w, id) {
		return
	}
	task, err := api.store.GetTask(r.Context(), id)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func (api *taskAPI) deleteTask(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if !checkID(w, id) {
		return
	}
	if err := api.store.DeleteTask(r.Context(), id); err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, struct{}{})
}

func (api *taskAPI) completeTask(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if !checkID(w, id) {
		return
	}
	err := api.store.CompleteTask(r.Context(), id, func(date, rule string) (string, error) {
		return NextDate(time.Now(), date, rule)
	})
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, struct{}{})
}
