package kit

import (
	"os"

	"github.com/PondWader/kit/pkg/db"
)

const Version = "0.0.1"

func New() (*Kit, error) {
	k := Kit{}
	if err := k.setupHome(); err != nil {
		return nil, err
	}

	if err := k.loadRepos(); err != nil {
		return nil, err
	}

	return &k, nil
}

type Kit struct {
	Home  *os.Root
	DB    *db.DB
	Repos []Repo
}

func (k *Kit) Close() error {
	err1 := k.DB.Close()
	err2 := k.Home.Close()
	if err1 != nil {
		return err1
	}
	return err2
}
