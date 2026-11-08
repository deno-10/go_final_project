package db

import (
	"context"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func TestConcurrentCompletionsUseLatestDate(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "scheduler.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	ctx := context.Background()
	id, err := store.AddTask(ctx, &Task{Date: "20240101", Title: "Проверить почту", Repeat: "d 1"})
	if err != nil {
		t.Fatal(err)
	}
	taskID := strconv.FormatInt(id, 10)
	nextDay := func(date, _ string) (string, error) {
		day, err := time.Parse("20060102", date)
		return day.AddDate(0, 0, 1).Format("20060102"), err
	}

	const n = 8
	start := make(chan struct{})
	results := make(chan error, n)
	for i := 0; i < n; i++ {
		go func() {
			<-start
			results <- store.CompleteTask(ctx, taskID, nextDay)
		}()
	}
	close(start)
	for i := 0; i < n; i++ {
		if err := <-results; err != nil {
			t.Error(err)
		}
	}
	task, err := store.GetTask(ctx, taskID)
	if err != nil {
		t.Fatal(err)
	}
	if task.Date != "20240109" {
		t.Errorf("date after %d completions = %s, want 20240109", n, task.Date)
	}
}
