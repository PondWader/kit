package kit

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/PondWader/kit/pkg/lang"
	"github.com/PondWader/kit/pkg/lang/values"
)

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

	slices.SortFunc(versions, compareVersions)

	return versions, nil
}

func (p *Package) Install(version string) error {
	env, err := p.loadEnv()
	if err != nil {
		return err
	}

	installV, err := env.GetExport("install")
	if err != nil {
		return err
	}
	installFn, ok := installV.ToFunction()
	if !ok {
		return fmt.Errorf("error running install in %s: expected install export to be a function", filepath.Join(p.Path, "package.kit"))
	}

	sb := &installBinding{}
	sb.Load(env)

	_, cErr := installFn.Call(values.String(version).Val())
	if cErr != nil {
		return cErr
	}
	return nil
}

func compareVersions(a, b string) int {
	partsA := strings.Split(a, ".")
	partsB := strings.Split(b, ".")

	maxLen := max(len(partsB), len(partsA))

	for i := range maxLen {
		var partA, partB string
		if i < len(partsA) {
			partA = partsA[i]
		}
		if i < len(partsB) {
			partB = partsB[i]
		}

		if cmp := compareVersionPart(partA, partB); cmp != 0 {
			return cmp
		}
	}

	return 0
}

func compareVersionPart(a, b string) int {
	numA, suffixA := parseVersionPart(a)
	numB, suffixB := parseVersionPart(b)

	if numA != numB {
		return numA - numB
	}

	// No suffix (release) is greater than any pre-release suffix
	if suffixA == "" && suffixB != "" {
		return 1
	}
	if suffixA != "" && suffixB == "" {
		return -1
	}

	return strings.Compare(suffixA, suffixB)
}

func parseVersionPart(part string) (num int, suffix string) {
	if part == "" {
		return 0, ""
	}

	i := 0
	for i < len(part) && part[i] >= '0' && part[i] <= '9' {
		num = num*10 + int(part[i]-'0')
		i++
	}

	suffix = part[i:]
	return num, suffix
}
