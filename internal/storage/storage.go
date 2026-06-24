package storage

import (
	"DeadLinkChecker/internal/scrapper"
	"context"
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
		panic("db open error")
	}
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("can't connect db: %w", err)
	}

	return &Storage{db: db}, nil

}

func (s *Storage) Save(ctx context.Context, p *scrapper.Page) error {

	q := `INSERT INTO table (url, isDead) VALUES (?, ?)`

	if _, err := s.db.ExecContext(ctx, q, p.URL, p.IsDead); err != nil {
		return fmt.Errorf("cant save page: %w", err)
	}

	return nil
}

func (s *Storage) ClearDB() error {
	q := `DELETE FROM table;`

	if _, err := s.db.Exec(q); err != nil {
		return fmt.Errorf("Delete table error: %w", err)
	}

	return nil
}
