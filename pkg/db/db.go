package db

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

type DB struct {
	sql *sql.DB
}

func Open(path string) (*DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if _, err = db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		return nil, err
	}

	wrapper := &DB{db}
	if err = wrapper.applyMigrations(); err != nil {
		return nil, err
	}
	return wrapper, nil
}

func (db *DB) Close() error {
	return db.sql.Close()
}
