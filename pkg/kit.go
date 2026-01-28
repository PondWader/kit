package kit

import (
	"fmt"
	"path/filepath"
	"slices"

	"github.com/PondWader/kit/pkg/db"
	"github.com/PondWader/kit/pkg/lang"
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
	Home  KitFS
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

type Package struct {
	Name string
	Path string
	Repo string

	k   *Kit
	env *lang.Environment
}

func (p *Package) loadEnv() (*lang.Environment, error) {
	if p.env != nil {
		return p.env, nil
	}

	f, err := p.k.Home.Open(filepath.Join(p.Path, "package.kit"))
	if err != nil {
		return nil, err
	}
	env, err := lang.Execute(f)
	if err != nil {
		return nil, err
	}
	env.LoadStd()
	return env, nil
}

func (p *Package) Versions() ([]string, error) {
	env, err := p.loadEnv()
	if err != nil {
		return nil, err
	}

	versionsV, err := env.GetExport("versions")
	if err != nil {
		return nil, err
	}
	versionsFn, ok := versionsV.ToFunction()
	if !ok {
		return nil, fmt.Errorf("error getting versions from %s: expected versions export to be a function", filepath.Join(p.Path, "package.kit"))
	}

	returned, cErr := versionsFn.Call()
	if cErr != nil {
		return nil, cErr
	}
	versionsList, ok := returned.ToList()
	if !ok {
		return nil, fmt.Errorf("error getting versions from %s: expected versions export return type to be a list", filepath.Join(p.Path, "package.kit"))
	}

	versions := make([]string, 0, versionsList.Size())
	foundVersions := make(map[string]struct{})

	for _, v := range versionsList.AsSlice() {
		vStr, ok := v.ToString()
		if !ok {
			return nil, fmt.Errorf("error getting versions from %s: expected versions element to be a string", filepath.Join(p.Path, "package.kit"))
		}
		ver := vStr.String()
		if _, ok := foundVersions[ver]; ok {
			continue
		}
		foundVersions[ver] = struct{}{}

		versions = append(versions, ver)
	}

	slices.Sort(versions)

	return versions, nil
}
