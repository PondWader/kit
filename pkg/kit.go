package kit

import (
	"github.com/PondWader/kit/pkg/db"
)

const Version = "0.0.1"

func New(autoPull bool) (*Kit, error) {
	k := Kit{autoPull: autoPull}
	if err := k.setupHome(); err != nil {
		return nil, err
	}

	if err := k.loadRepos(); err != nil {
		return nil, err
	}

	return &k, nil
}

type Kit struct {
	Home     KitFS
	DB       *db.DB
	Repos    []Repo
	autoPull bool
}

func (k *Kit) Close() error {
	err1 := k.DB.Close()
	err2 := k.Home.Close()
	if err1 != nil {
		return err1
	}
	return err2
}

func (k *Kit) LoadPackage(name string) ([]*Package, error) {
	pkgsInfo, err := k.DB.GetPackages(name)
	if err != nil {
		return nil, err
	}

	pkgs := make([]*Package, len(pkgsInfo))
	for i, pkgInfo := range pkgsInfo {
		pkgs[i] = &Package{
			Name: pkgInfo.Name,
			Path: pkgInfo.Path,
			Repo: pkgInfo.Repo,

			k: k,
		}
	}

	return pkgs, nil
}
