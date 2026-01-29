package main

import (
	"flag"
	"os"

	"github.com/PondWader/kit/internal/render"
	kit "github.com/PondWader/kit/pkg"
)

var PullCommand = Command{
	Name:        "pull",
	Description: " pulls the latest version of all repositories",
	Run: func(fs *flag.FlagSet) {
		t := render.NewTerm(os.Stdin, os.Stdout)
		defer t.Stop()

		k, err := kit.New(false, t)
		if err != nil {
			printError(err)
			os.Exit(1)
		}

		if err = k.PullRepos(); err != nil {
			printError(err)
			os.Exit(1)
		}
	},
}
