package db

import (
	"database/sql"
	"errors"
	"time"

	_ "modernc.org/sqlite"
)

var ErrNoData = errors.New("no data available")

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

type CoreInfo struct {
	LastPulledAt      time.Time
	LastPullRepoMtime time.Time
}

func (db *DB) GetCoreInfo() (CoreInfo, error) {
	var id int
	var i CoreInfo

	var lastPulledAtRaw string
	var lastPullRepoMtimeRaw string

	row := db.sql.QueryRow("SELECT * FROM pull_info;")
	err := row.Scan(&id, &lastPulledAtRaw, &lastPullRepoMtimeRaw)
	if err == sql.ErrNoRows {
		return i, ErrNoData
	}

	if i.LastPulledAt, err = time.Parse(time.RFC3339, lastPulledAtRaw); err != nil {
		return i, err
	}
	i.LastPullRepoMtime, err = time.Parse(time.RFC3339, lastPullRepoMtimeRaw)
	return i, nil
}

func (db *DB) UpdateCoreInfo(i CoreInfo) error {
	_, err := db.sql.Exec(`INSERT INTO pull_info (id, last_pulled_at, last_pull_list_mtime)
		VALUES (1, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			last_pulled_at = excluded.last_pulled_at,
			last_pull_list_mtime = excluded.last_pull_list_mtime;`,
		i.LastPulledAt.Format(time.RFC3339), i.LastPullRepoMtime.Format(time.RFC3339))
	return err
}

func (db *DB) BeginPackageIndex(repo string) (*PackageIndex, error) {
	tx, err := db.sql.Begin()
	if err != nil {
		return nil, err
	}

	if _, err = tx.Exec("DELETE FROM packages WHERE repo = ?;", repo); err != nil {
		return nil, err
	}

	return &PackageIndex{tx, repo}, nil
}

type PackageInfo struct {
	Name string
	Repo string
	Path string
}

func (db *DB) GetPackages(name string) ([]PackageInfo, error) {
	rows, err := db.sql.Query("SELECT name, repo, path FROM packages WHERE name = ?", name)
	if err != nil {
		return nil, err
	}

	var pkgs []PackageInfo
	for rows.Next() {
		var pkg PackageInfo
		if err = rows.Scan(&pkg.Name, &pkg.Repo, &pkg.Path); err != nil {
			return nil, err
		}
		pkgs = append(pkgs, pkg)
	}
	return pkgs, nil
}

type PackageIndex struct {
	tx   *sql.Tx
	repo string
}

func (i *PackageIndex) Rollback() error {
	return i.tx.Rollback()
}

func (i *PackageIndex) Commit() error {
	return i.tx.Commit()
}

func (i *PackageIndex) IndexPackage(name, path string) error {
	_, err := i.tx.Exec("INSERT INTO packages VALUES (?, ?, ?);", name, i.repo, path)
	return err
}
