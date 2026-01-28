package db

import (
	"slices"
	"time"

	"github.com/PondWader/kit/include"
)

func (db *DB) applyMigrations() (err error) {
	if _, err := db.sql.Exec(`CREATE TABLE IF NOT EXISTS migrations (
		name TEXT PRIMARY KEY UNIQUE,
		applied_at DATETIME
	);`); err != nil {
		return err
	}

	entries, err := include.Migrations.ReadDir("migrations")
	if err != nil {
		return err
	}

	// Begin transaction
	tx, err := db.sql.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	rows, err := tx.Query("SELECT name FROM migrations;")
	if err != nil {
		return err
	}

	var names []string
	for rows.Next() {
		var name string
		rows.Scan(&name)
		names = append(names, name)
	}

	for _, entry := range entries {
		if !slices.Contains(names, entry.Name()) {
			// Apply the migration if it has not already been applied
			migration, err := include.Migrations.ReadFile("migrations/" + entry.Name())
			if err != nil {
				return err
			}
			if _, err = tx.Exec(string(migration)); err != nil {
				return err
			}

			tx.Exec("INSERT INTO migrations VALUES (?, ?);", entry.Name(), time.Now().Format(time.RFC3339))
		}
	}

	// End transaction
	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}
