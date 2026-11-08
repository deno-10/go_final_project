package db

import (
	"context"
	"database/sql"
	"time"
)

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

const taskColumns = "id, date, title, COALESCE(comment, ''), COALESCE(repeat, '')"

func (s *Store) AddTask(ctx context.Context, task *Task) (int64, error) {
	result, err := s.db.ExecContext(ctx,
		"INSERT INTO scheduler(date, title, comment, repeat) VALUES (?, ?, ?, ?)",
		task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (s *Store) GetTask(ctx context.Context, id string) (*Task, error) {
	var task Task
	row := s.db.QueryRowContext(ctx, "SELECT "+taskColumns+" FROM scheduler WHERE id = ?", id)
	if err := row.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
		return nil, err
	}
	return &task, nil
}

func (s *Store) Tasks(ctx context.Context, search string, limit int) ([]Task, error) {
	query := "SELECT " + taskColumns + " FROM scheduler"
	var args []any
	if search != "" {
		if date, err := time.Parse("02.01.2006", search); err == nil {
			query += " WHERE date = ?"
			args = append(args, date.Format("20060102"))
		} else {
			query += " WHERE instr(title, ?) > 0 OR instr(comment, ?) > 0"
			args = append(args, search, search)
		}
	}
	query += " ORDER BY date, id LIMIT ?"
	rows, err := s.db.QueryContext(ctx, query, append(args, limit)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []Task{}
	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
}

func (s *Store) UpdateTask(ctx context.Context, task *Task) error {
	return checkAffected(s.db.ExecContext(ctx,
		"UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?",
		task.Date, task.Title, task.Comment, task.Repeat, task.ID))
}

func (s *Store) DeleteTask(ctx context.Context, id string) error {
	return checkAffected(s.db.ExecContext(ctx, "DELETE FROM scheduler WHERE id = ?", id))
}

func checkAffected(result sql.Result, err error) error {
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err == nil && n == 0 {
		return sql.ErrNoRows
	}
	return err
}

func (s *Store) CompleteTask(ctx context.Context, id string, next func(date, rule string) (string, error)) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var date, rule string
	row := tx.QueryRowContext(ctx, "SELECT date, COALESCE(repeat, '') FROM scheduler WHERE id = ?", id)
	if err := row.Scan(&date, &rule); err != nil {
		return err
	}
	if rule == "" {
		_, err = tx.ExecContext(ctx, "DELETE FROM scheduler WHERE id = ?", id)
	} else {
		nextDate, dateErr := next(date, rule)
		if dateErr != nil {
			return dateErr
		}
		_, err = tx.ExecContext(ctx, "UPDATE scheduler SET date = ? WHERE id = ?", nextDate, id)
	}
	if err != nil {
		return err
	}
	return tx.Commit()
}
