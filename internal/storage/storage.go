package storage

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

type DataBase struct {
	DB *sql.DB
}

func (db *DataBase) ConnectDB() (*sql.DB, error) {

	database, err := sql.Open("sqlite", "stringurl")
	if err != nil {
		panic("Connect db error")
	}

	return database, nil
}
