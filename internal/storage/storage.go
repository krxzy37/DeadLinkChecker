package storage

import (
	"DeadLinkChecker/internal/scrapper"
	"database/sql"
	"fmt"
	"net/url"

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

	q := `INSERT INTO pages (url, status_code) VALUES (?, ?)`

	if _, err := s.db.Exec(q, p.URL, p.StatusCode); err != nil {
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
									status_code INTEGER
									)`

	if _, err := s.db.Exec(q); err != nil {
		return fmt.Errorf("Cant create table: %w", err)
	}

	return nil
}

func (s *Storage) Close() {
	s.db.Close()
}

func (s *Storage) GetPages(primalLink string) ([]scrapper.Page, error) {

	Pages := []scrapper.Page{}

	base, err := url.Parse(primalLink)
	if err != nil {
		return nil, fmt.Errorf("cant parse primal link: %w", err)
	}

	pattern := "%" + base.Host + "%"

	rows, err := s.db.Query("SELECT url, status_code FROM pages WHERE url LIKE ?", pattern)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var page scrapper.Page

		if err = rows.Scan(&page.URL, &page.StatusCode); err != nil {
			return nil, fmt.Errorf("cant add in slice page from db: %w", err)
		}
		Pages = append(Pages, page)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error while iteration: %w", err)
	}

	return Pages, nil

}

func (s *Storage) ChekPage(chekLink string) (bool, error) {

	pattern := "%" + chekLink + "%"
	var exists bool

	q := `SELECT EXISTS(SELECT 1 FROM pages WHERE url LIKE ?)`

	if err := s.db.QueryRow(q, pattern).Scan(&exists); err != nil {
		return false, fmt.Errorf("cant find link: %w", err)
	}

	return exists, nil
}
