package kit

import (
	"database/sql"
	"os"
	"path/filepath"
	"slices"

	"github.com/PondWader/kit/include"
	_ "modernc.org/sqlite"
)

func (k *Kit) setupHome() error {
	entries, err := os.ReadDir(k.Home.Name())
	if err != nil {
		return err
	}

	// Make all the missing directories
	dirs := [4]string{"bin", "lib", "repos", "packages"}
	for _, dir := range dirs {
		if !slices.ContainsFunc(entries, func(e os.DirEntry) bool {
			return e.Name() == dir
		}) {
			if err := k.Home.Mkdir(dir, 0755); err != nil {
				return err
			}
		}
	}

	// Check for missing "repos.kit" to specify the repositories and add it if it doesn't exist
	if !slices.ContainsFunc(entries, func(e os.DirEntry) bool {
		return e.Name() == "repos.kit"
	}) {
		f, err := k.Home.OpenFile("repos.kit", os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return err
		}
		_, err = f.WriteString(include.Repositories)
		if err != nil {
			return err
		}
	}

	// Open the SQLite DB
	db, err := sql.Open("sqlite", filepath.Join(k.Home.Name(), "kit.sqlite"))
	if err != nil {
		return err
	}
	k.DB = db

	return nil
}
