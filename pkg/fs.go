package kit

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"

	"github.com/PondWader/kit/include"
	"github.com/PondWader/kit/pkg/db"
	_ "modernc.org/sqlite"
)

func (k *Kit) setupHome() error {
	// Resolve the home directory
	home, err := resolveHome()
	if err != nil {
		return err
	}
	if err = os.MkdirAll(home, 0755); err != nil {
		return err
	}
	root, err := os.OpenRoot(home)
	if err != nil {
		return err
	}
	k.Home = root

	entries, err := os.ReadDir(root.Name())
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
	db, err := db.Open(filepath.Join(k.Home.Name(), "kit.sqlite"))
	if err != nil {
		return err
	}
	k.DB = db

	return nil
}

func resolveHome() (string, error) {
	// Ideally KIT_HOME should be set
	home := os.Getenv("KIT_HOME")
	if home != "" {
		return home, nil
	}

	// If not, fallback to using the user's data home
	dataHome, err := resolveDataHome()
	if err != nil {
		return "", err
	}
	return filepath.Join(dataHome, "kit"), nil
}

func resolveDataHome() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	switch runtime.GOOS {
	case "windows":
		dataHome := os.Getenv("LOCALAPPDATA")
		if dataHome != "" {
			return dataHome, nil
		}
		return filepath.Join(home, "./AppData/Local"), nil
	case "darwin":
		return filepath.Join(home, "./Library"), nil
	default:
		dataHome := os.Getenv("XDG_DATA_HOME")
		if dataHome != "" {
			return dataHome, nil
		}
		return filepath.Join(home, "./.local/share"), nil
	}
}
