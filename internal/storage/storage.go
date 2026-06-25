package storage

import (
	"DeadLinkChecker/internal/scrapper"
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

type Storage struct {
	db *sql.DB
}

func New(path string) (*Storage, error) {

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("sql open error: %w", err)
	}
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("can't connect db: %w", err)
	}

	return &Storage{db: db}, nil

}

func (s *Storage) Save(p scrapper.Page) error {

	q := `INSERT INTO pages (url, isDead) VALUES (?, ?)`

	if _, err := s.db.Exec(q, p.URL, p.IsDead); err != nil {
		return fmt.Errorf("cant save page: %w", err)
	}

	return nil
}

func (s *Storage) ClearDB() error {
	q := `DELETE FROM pages;`

	if _, err := s.db.Exec(q); err != nil {
		return fmt.Errorf("Delete table error: %w", err)
	}

	return nil
}

func (s *Storage) Init() error {

	q := `CREATE TABLE IF NOT EXISTS pages (
									id INTEGER PRIMARY KEY AUTOINCREMENT,
									url TEXT NOT NULL, 
									is_dead BOOLEAN NOT NULL DEFAULT 0 CHECK (is_dead IN (0, 1))
									)`

	if _, err := s.db.Exec(q); err != nil {
		return fmt.Errorf("Cant create table: %w", err)
	}

	return nil
}

func (s *Storage) Close() {
	s.db.Close()
}
