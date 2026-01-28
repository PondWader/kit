package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	kit "github.com/PondWader/kit/pkg"
)

var VersionsCommand = Command{
	Name:             "versions",
	Usage:            "<package>",
	Description:      "lists all versions available for a package",
	RequiredArgCount: 1,
	Run: func(fs *flag.FlagSet) {
		start := time.Now()
		pkgName := fs.Arg(0)
		k, err := kit.New(true)
		if err != nil {
			printError(err)
			os.Exit(1)
		}

		pkg := getPkg(k, pkgName)

		versions, err := pkg.Versions()
		if err != nil {
			printError(err)
			os.Exit(1)
		}

		fmt.Println(strings.Join(versions, "\n"))

		fmt.Println("completed in", time.Since(start))
	},
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
