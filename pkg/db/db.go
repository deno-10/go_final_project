package db

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

const schema = `
CREATE TABLE IF NOT EXISTS scheduler (
	id      INTEGER PRIMARY KEY AUTOINCREMENT,
	date    CHAR(8) NOT NULL CHECK (length(date) = 8 AND date NOT GLOB '*[^0-9]*'),
	title   VARCHAR NOT NULL,
	comment TEXT NOT NULL DEFAULT '',
	repeat  VARCHAR(128) NOT NULL DEFAULT '' CHECK (length(repeat) <= 128)
);
CREATE INDEX IF NOT EXISTS scheduler_date_idx ON scheduler(date);`

type Store struct {
	db *sql.DB
}

func Open(path string) (*Store, error) {
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	// SQLite permits one writer at a time. Queue requests on the same connection.
	conn.SetMaxOpenConns(1)
	if _, err := conn.Exec("PRAGMA busy_timeout = 5000"); err != nil {
		conn.Close()
		return nil, err
	}
	if _, err := conn.Exec(schema); err != nil {
		conn.Close()
		return nil, err
	}
	return &Store{db: conn}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}
