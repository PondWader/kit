package kit

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/PondWader/kit/internal/render"
	"github.com/PondWader/kit/pkg/db"
	"github.com/PondWader/kit/pkg/lang"
	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
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
	defer reposFile.Close()

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

		if slices.ContainsFunc(repos, func(r Repo) bool {
			return r.Name == repo.Name
		}) {
			return fmt.Errorf("error loading %s: name \"%s\" is duplicated", filepath.Join(k.Home.Name(), "repos.kit"), repo.Name)
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
	s := render.NewSpinner("Pulling repositories...")
	defer s.Stop()

	r := render.NewRenderer(os.Stdout)
	r.Mount(s)

	time.Sleep(4 * time.Second) // Run for some time to simulate work

	// dirs, err := k.repoDirs()
	// if err != nil {
	// 	return fmt.Errorf("error pulling repositories: %w", err)
	// }

	for _, repo := range k.Repos {
		if repo.Type != "git" {
			return errors.New("error pulling repos: repository type \"" + repo.Type + "\" is not supported (only \"git\" is supported at this time)")
		}
		_, err := git.PlainClone(filepath.Join(k.Home.Name(), "repos", repo.Name), &git.CloneOptions{
			URL:           repo.URL,
			ReferenceName: plumbing.ReferenceName(repo.Branch),
			SingleBranch:  true,
			Depth:         0,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
