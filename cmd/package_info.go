package main

import (
	"flag"
	"fmt"
	"os"
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
		k, err := kit.New()
		if err != nil {
			printError(err)
			os.Exit(1)
		}

		_ = k
		_ = pkgName
		fmt.Println("completed in", time.Since(start))
	},
}
