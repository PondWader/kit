package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/PondWader/kit/internal/ansi"
	"github.com/PondWader/kit/internal/render"
	kit "github.com/PondWader/kit/pkg"
)

var VersionsCommand = Command{
	Name:             "versions",
	Usage:            "<package>",
	Description:      "lists all versions available for a package",
	RequiredArgCount: 1,
	Run: func(fs *flag.FlagSet) {
		t := render.NewTerm(os.Stdin, os.Stdout)
		defer t.Stop()

		pkgName := fs.Arg(0)
		k, err := kit.New(true, t)
		if err != nil {
			printError(err)
			os.Exit(1)
		}

		pkg := getPkg(k, pkgName)

		s := render.NewSpinner("Fetching versions...")
		t.Mount(s)

		versions, err := pkg.Versions()
		if err != nil {
			s.Stop()
			printError(err)
			os.Exit(1)
		}

		s.Stop()

		fmt.Println(strings.Join(versions, "\n"))
	},
	TaskRunner: true,
}

var InstallCommand = Command{
	Aliases:          []string{"get"},
	Name:             "install",
	Usage:            "<package> [version]",
	Description:      "install a package",
	RequiredArgCount: 1,
	OptionalArgCount: 2,
	Run: func(fs *flag.FlagSet) {
		t := render.NewTerm(os.Stdin, os.Stdout)
		defer t.Stop()

		pkgName := fs.Arg(0)
		k, err := kit.New(true, t)
		if err != nil {
			printError(err)
			os.Exit(1)
		}

		versionSpec := "latest"
		if fs.NArg() > 1 {
			versionSpec = fs.Arg(1)
		}

		pkg := getPkg(k, pkgName)

		s := render.NewSpinner(fmt.Sprintf("Installing %s"+ansi.BrightBlue("@")+"%s...", ansi.Cyan(pkgName), ansi.Cyan(versionSpec)))
		t.Mount(s)

		time.Sleep(time.Second * 2)
		versions, err := pkg.Versions()
		if err != nil {
			s.Stop()
			printError(err)
			os.Exit(1)
		}

		var version string
		if versionSpec == "latest" {
			version = pickLatestVersion(versions)
		} else if !slices.Contains(versions, versionSpec) {
			var ok bool
			version, ok = matchVersion(versionSpec, versions)
			if !ok {
				printError(errors.New("could not match version: " + versionSpec))
				os.Exit(1)
			}
		}

		if err = pkg.Install(version); err != nil {
			s.Stop()
			printError(err)
			os.Exit(1)
		}

		s.Succeed(fmt.Sprintf("Installed %s"+ansi.BrightBlue("@")+"%s", ansi.Cyan(pkgName), ansi.Cyan(version)))
	},
	TaskRunner: true,
}

func getPkg(k *kit.Kit, name string) *kit.Package {
	pkgs, err := k.LoadPackage(name)
	if err != nil {
		printError(err)
		os.Exit(1)
	} else if len(pkgs) == 0 {
		printError(errors.New("no packages found matching name"))
		os.Exit(1)
	}
	// TODO: Ask user to select a package if there are multiple
	return pkgs[0]
}

func pickLatestVersion(versions []string) string {
	for i := len(versions) - 1; i >= 0; i-- {
		// Pick last version without letters in it (e.g. skip 1.26rc2) if possible
		if !hasLetters(versions[i]) {
			return versions[i]
		}
	}
	return versions[len(versions)-1]
}

func matchVersion(versionSpec string, versions []string) (string, bool) {
	for i := len(versions) - 1; i >= 0; i-- {
		version := versions[i]
		// Exact match
		if version == versionSpec {
			return version, true
		}
		// Prefix match (e.g., "1.26" matches "1.26.8")
		if strings.HasPrefix(version, versionSpec+".") {
			// Skip versions with letters (e.g., "1.26rc2")
			if !hasLetters(version) {
				return version, true
			}
		}
	}
	return "", false
}

func hasLetters(str string) bool {
	for _, c := range str {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
			return true
		}
	}
	return false
}
