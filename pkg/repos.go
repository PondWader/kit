package kit

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/PondWader/kit/pkg/db"
	"github.com/PondWader/kit/pkg/lang"
)

type Repo struct {
	Name   string
	Type   string
	URL    string
	Branch string
	Dir    string
}

func (k *Kit) loadRepos() error {
	reposFile, err := k.Home.Open("repos.kit")
	if err != nil {
		return err
	}

	env, err := lang.Execute(reposFile)
	if err != nil {
		return fmt.Errorf("error loading %s: %w", filepath.Join(k.Home.Name(), "repos.kit"), err)
	}
	reposV, err := env.GetExport("repositories")
	if err != nil {
		return fmt.Errorf("error loading %s: %w", filepath.Join(k.Home.Name(), "repos.kit"), err)
	}

	l, ok := reposV.ToList()
	if !ok {
		return fmt.Errorf("error loading %s: expected \"repositories\" export to be a list", filepath.Join(k.Home.Name(), "repos.kit"))
	}

	repos := make([]Repo, l.Size())

	for i, repoV := range l.AsSlice() {
		var repo Repo

		o, ok := repoV.ToObject()
		if !ok {
			return fmt.Errorf("error loading %s: expected repository item to be an object", filepath.Join(k.Home.Name(), "repos.kit"))
		}
		repo.Name, err = o.GetString("name")
		if err != nil {
			return fmt.Errorf("error loading %s: %w", filepath.Join(k.Home.Name(), "repos.kit"), err)
		}
		repo.Type, err = o.GetString("type")
		if err != nil {
			return fmt.Errorf("error loading %s: %w", filepath.Join(k.Home.Name(), "repos.kit"), err)
		}
		repo.URL, err = o.GetString("url")
		if err != nil {
			return fmt.Errorf("error loading %s: %w", filepath.Join(k.Home.Name(), "repos.kit"), err)
		}
		repo.Branch, err = o.GetString("branch")
		if err != nil {
			return fmt.Errorf("error loading %s: %w", filepath.Join(k.Home.Name(), "repos.kit"), err)
		}
		repo.Dir, err = o.GetString("dir")
		if err != nil {
			return fmt.Errorf("error loading %s: %w", filepath.Join(k.Home.Name(), "repos.kit"), err)
		}

		repos[i] = repo
	}

	k.Repos = repos

	return k.checkForAutoRepoPull()
}

func (k *Kit) checkForAutoRepoPull() error {
	info, err := k.DB.GetCoreInfo()
	if err != nil && err != db.ErrNoData {
		return err
	}

	finfo, err := k.Home.Stat("repos.kit")
	if err != nil && err != db.ErrNoData {
		return err
	}

	// If the file has not changed or the last pull was less than 24 hours a day, don't do an auto pull
	if finfo.ModTime().Equal(info.LastPullRepoMtime) && time.Since(info.LastPulledAt) < time.Hour*24 {
		return nil
	}

	return k.PullRepos()
}

func (k *Kit) PullRepos() error {
	fmt.Println("Pulling repos...")
	return nil
}
