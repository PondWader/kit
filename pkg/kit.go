package kit

import (
	"database/sql"
	"os"
	"path/filepath"
	"runtime"
)

const Version = "0.0.1"

func New() (*Kit, error) {
	home, err := resolveHome()
	if err != nil {
		return nil, err
	}
	root, err := os.OpenRoot(home)
	if err != nil {
		return nil, err
	}

	k := Kit{Home: root}
	k.setupHome()

	return &k, nil
}

type Kit struct {
	Home *os.Root
	DB   *sql.DB
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
