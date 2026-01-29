package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

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

		_ = k
		_ = pkgName
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
