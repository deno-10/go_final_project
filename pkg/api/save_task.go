package api

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"scheduler/pkg/db"
)

func readTask(w http.ResponseWriter, r *http.Request) (*db.Task, error) {
	var task db.Task
	if err := decodeJSON(w, r, &task); err != nil {
		return nil, err
	}
	if strings.TrimSpace(task.Title) == "" {
		return nil, errors.New("Не указан заголовок задачи")
	}
	if len(task.Repeat) > 128 {
		return nil, errors.New("Правило повторения длиннее 128 символов")
	}

	now := time.Now()
	today := now.Format(dateLayout)
	if task.Date == "" {
		task.Date = today
	}
	if _, err := time.Parse(dateLayout, task.Date); err != nil {
		return nil, errors.New("Дата должна существовать и иметь формат ГГГГММДД")
	}

	next := today
	if task.Repeat != "" {
		var err error
		next, err = NextDate(now, task.Date, task.Repeat)
		if err != nil {
			return nil, err
		}
	}
	if task.Date < today {
		task.Date = next
	}
	return &task, nil
}

func (api *taskAPI) addTask(w http.ResponseWriter, r *http.Request) {
	task, err := readTask(w, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	id, err := api.store.AddTask(r.Context(), task)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"id": strconv.FormatInt(id, 10)})
}

func (api *taskAPI) editTask(w http.ResponseWriter, r *http.Request) {
	task, err := readTask(w, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if !checkID(w, task.ID) {
		return
	}
	if err := api.store.UpdateTask(r.Context(), task); err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, struct{}{})
}
