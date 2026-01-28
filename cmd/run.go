package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/PondWader/kit/internal/ansi"
)

type Command struct {
	Name             string
	Usage            string
	Description      string
	Flags            *flag.FlagSet
	RequiredArgCount int
	OptionalArgCount int
	Run              func(fs *flag.FlagSet)
	Aliases          []string
	Hidden           bool
}

var Commands = []Command{
	HelpCommand,
	VersionCommand,
	VersionsCommand,
	PullCommand,
}

func main() {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	displayVersion := fs.Bool("version", false, "Displays the version")

	if err := fs.Parse(os.Args[1:]); err != nil {
		if err == flag.ErrHelp {
			HelpCommand.Run(nil)
			return
		}
		printError(err)
		os.Exit(1)
	}

	if *displayVersion {
		VersionCommand.Run(nil)
		return
	} else if fs.NArg() == 0 {
		HelpCommand.Run(nil)
		return
	}

	subcmd := fs.Arg(0)

	for _, cmd := range Commands {
		if cmd.Name == subcmd || slices.Contains(cmd.Aliases, subcmd) {
			flags := cmd.Flags
			if flags == nil {
				flags = flag.NewFlagSet("", flag.ContinueOnError)
			}
			if err := flags.Parse(fs.Args()[1:]); err != nil && err != flag.ErrHelp {
				printError(err)
				os.Exit(1)
			}
			if flags.NArg() < cmd.RequiredArgCount {
				printError(errors.New("missing arguments! Correct usage: " + cmd.Name + " " + cmd.Usage))
				os.Exit(1)
			} else if flags.NArg() > cmd.RequiredArgCount+cmd.OptionalArgCount {
				printError(errors.New("too many arguments! Correct usage: " + cmd.Name + " " + cmd.Usage))
				os.Exit(1)
			}

			cmd.Run(
				flags,
			)
			return
		}
	}

	printError(errors.New("no matching command found for \"" + subcmd + "\""))

}

func printError(err error) {
	msg := err.Error()
	fmt.Println(ansi.Bold(ansi.Red("ERROR ")) + strings.ToUpper(string(msg[0])) + msg[1:])
}
