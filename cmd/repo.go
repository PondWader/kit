package main

import (
	"flag"
	"os"

	kit "github.com/PondWader/kit/pkg"
)

var PullCommand = Command{
	Name:        "pull",
	Description: " pulls the latest version of all repositories",
	Run: func(fs *flag.FlagSet) {
		k, err := kit.New(false)
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
