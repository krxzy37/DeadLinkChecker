package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func ConnectDB() (*sql.DB, error) {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("файл .env не найдет")
	}

	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPass, dbHost, dbPort, dbName)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("ошибка инициализации дб: %v", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("база недоступна (ping error): %w", err)
	}

	return db, nil
}

func SaveResult(db *sql.DB, link string, isBroken bool, errMsg string, connected []string) error {

	jsonConnected, erd := json.Marshal(connected)
	if erd != nil {
		return erd
	}

	query := `
				INSERT INTO visited_links (url, is_broken, error_msg, connected_links)
				VALUES ($1, $2, $3, $4)
				ON CONFLICT (url) DO NOTHING;
				`

	_, err := db.Exec(query, link, isBroken, errMsg, jsonConnected)
	if err != nil {
		return fmt.Errorf("ошибка записи в бд: %w", err)
	}

	return nil
}
