package kit

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/PondWader/kit/internal/gitcli"
	"github.com/PondWader/kit/internal/render"
	"github.com/PondWader/kit/pkg/db"
	"github.com/PondWader/kit/pkg/lang"
	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/transport"
	"github.com/go-git/go-git/v6/plumbing/transport/http"
)

type Repo struct {
	Name   string
	Type   string
	URL    string
	Branch string
	Dir    string
}

func (k *Kit) loadRepos() error {
	reposFile, err := k.Home.Open("repositories.kit")
	if err != nil {
		return err
	}
	defer reposFile.Close()

	env, err := lang.Execute(reposFile)
	if err != nil {
		return fmt.Errorf("error loading %s: %w", filepath.Join(k.Home.Name(), "repositories.kit"), err)
	}
	reposV, err := env.GetExport("repositories")
	if err != nil {
		return fmt.Errorf("error loading %s: %w", filepath.Join(k.Home.Name(), "repositories.kit"), err)
	}

	l, ok := reposV.ToList()
	if !ok {
		return fmt.Errorf("error loading %s: expected \"repositories\" export to be a list", filepath.Join(k.Home.Name(), "repositories.kit"))
	}

	repos := make([]Repo, l.Size())

	for i, repoV := range l.AsSlice() {
		var repo Repo

		o, ok := repoV.ToObject()
		if !ok {
			return fmt.Errorf("error loading %s: expected repository item to be an object", filepath.Join(k.Home.Name(), "repositories.kit"))
		}
		repo.Name, err = o.GetString("name")
		if err != nil {
			return fmt.Errorf("error loading %s: %w", filepath.Join(k.Home.Name(), "repositories.kit"), err)
		}
		repo.Type, err = o.GetString("type")
		if err != nil {
			return fmt.Errorf("error loading %s: %w", filepath.Join(k.Home.Name(), "repositories.kit"), err)
		}
		repo.URL, err = o.GetString("url")
		if err != nil {
			return fmt.Errorf("error loading %s: %w", filepath.Join(k.Home.Name(), "repositories.kit"), err)
		}
		repo.Branch, err = o.GetString("branch")
		if err != nil {
			return fmt.Errorf("error loading %s: %w", filepath.Join(k.Home.Name(), "repositories.kit"), err)
		}
		repo.Dir, err = o.GetString("dir")
		if err != nil {
			return fmt.Errorf("error loading %s: %w", filepath.Join(k.Home.Name(), "repositories.kit"), err)
		}

		if slices.ContainsFunc(repos, func(r Repo) bool {
			return r.Name == repo.Name
		}) {
			return fmt.Errorf("error loading %s: name \"%s\" is duplicated", filepath.Join(k.Home.Name(), "repositories.kit"), repo.Name)
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

	finfo, err := k.Home.Stat("repositories.kit")
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
	r := render.NewRenderer(os.Stdout)
	defer r.Stop()

	s := render.NewSpinner("Pulling repositories...")
	defer s.Stop()
	r.Mount(s)

	dirs, err := k.repoDirs()
	if err != nil {
		return fmt.Errorf("error pulling repositories: %w", err)
	}

	for _, repo := range k.Repos {
		if repo.Type != "git" {
			return errors.New("error pulling repos: repository type \"" + repo.Type + "\" is not supported (only \"git\" is supported at this time)")
		}

		// If it doesn't exist, have to clone it fresh
		if !slices.Contains(dirs, repo.Name) {
			cloneDir, err := os.MkdirTemp("", "kit_clone")
			if err != nil {
				return fmt.Errorf("error pulling repos: %w", err)
			}

			_, err = clone(cloneDir, &git.CloneOptions{
				URL:           repo.URL,
				ReferenceName: plumbing.ReferenceName(repo.Branch),
				SingleBranch:  true,
				Depth:         0,
			}, r)
			if err != nil {
				return err
			}

			targetDir := filepath.Join(k.Home.Name(), "repos", repo.Name)
			_ = targetDir
		}

	}

	return nil
}

func clone(path string, o *git.CloneOptions, r *render.Renderer) (*git.Repository, error) {
	repo, err := git.PlainClone(path, o)
	if err == nil {
		return repo, err
	} else if !errors.Is(err, transport.ErrAuthenticationRequired) {
		return repo, err
	} else if !strings.HasPrefix(o.URL, "https://") && !strings.HasPrefix(o.URL, "http://") {
		return repo, err
	}
	cloneErr := err

	if os.RemoveAll(path) != nil {
		return repo, cloneErr
	}

	// Try again with basic auth
	c := gitcli.Client{
		Prompt: func(prompt string, secret bool) (resp string, err error) {
			return "", errors.New("input prompts not supported yet")
		},
	}

	cred, err := c.GetCredentials(o.URL)
	if err != nil {
		return repo, cloneErr
	}
	o.Auth = &http.BasicAuth{
		Username: cred.Username,
		Password: cred.Password,
	}

	return git.PlainClone(path, o)
}
