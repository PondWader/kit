package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/PondWader/kit/internal/ansi"
	kit "github.com/PondWader/kit/pkg"
)

var HelpCommand = Command{
	Name: "help",
	Run: func(fs *flag.FlagSet) {
		printHelp()
	},
	Hidden: true,
}

var VersionCommand = Command{
	Name: "version",
	Run: func(fs *flag.FlagSet) {
		printVersion()
	},
	Hidden: true,
}

func printVersion() {
	fmt.Println(ansi.Color256(123, "Kit Package Manager ") + ansi.Bold(ansi.Cyan("v"+kit.Version)))
}

func printHelp() {
	fmt.Println("            \n" +
		"\x1b[38;5;202m     .=====.            \n" +
		"\x1b[38;5;208m   .=========.            \x1b[38;5;39m    _  ___ _ \x1b[0m\n" +
		"\x1b[38;5;214m  =============            \x1b[38;5;45m  | |/ (_) |_ \x1b[0m\n" +
		"\x1b[38;5;220m ===============            \x1b[38;5;51m | ' /| | __|\x1b[0m\n" +
		"\x1b[38;5;226m ===============            \x1b[38;5;87m | . \\| | |_ \x1b[0m\n" +
		"\x1b[38;5;220m ===============            \x1b[38;5;123m |_|\\_\\_|\\__|\x1b[0m\n" +
		"\x1b[38;5;214m  =============            \n" +
		"\x1b[38;5;208m   .=========.            \n" +
		"\x1b[38;5;202m     .=====.           \n" +
		"\x1b[38;5;130m    /   |   \\         \x1b[90mThe system package manager.\x1b[39m\n" +
		"\x1b[38;5;130m   /    |    \\            \n" +
		"\x1b[38;5;94m  [===========]            \n" +
		"\x1b[38;5;94m  |           |            \n" +
		"\x1b[38;5;94m  [===========]\x1b[0m            \n" +
		"           \n ")

	fmt.Println(fmtCommandMenu([]cmd{
		{Args: "install <package>[@version] (alias: add)", Desc: "install a package"},
		{Args: "uninstall <package>[@version] (alias: remove)", Desc: "uninstall a package"},
		{Args: "use <package>@<version>", Desc: "switch to a specific version of a package"},
		{Args: "list [repos/packages/available] (alias: ls)", Desc: "lists all repositories, installed packages or available packages (default: installed packages)"},
		{Args: "versions <package>", Desc: "lists all versions available for a package"},
		{Args: "search <term>", Desc: "search packages"},
		{Args: "pull", Desc: "pulls the latest version of all repositories"},
	}) + "\n")
}

type cmd struct {
	Args string
	Desc string
}

func fmtCommandMenu(cmds []cmd) string {
	var longestArgs int
	for _, cmd := range cmds {
		if len(cmd.Args) > longestArgs {
			longestArgs = len(cmd.Args)
		}
	}

	var sb strings.Builder
	for _, cmd := range cmds {
		fmt.Fprintf(&sb, "    %s", ansi.Color256(87, "kit "+cmd.Args))
		for range longestArgs - len(cmd.Args) + 5 {
			sb.WriteRune(' ')
		}
		sb.WriteString(ansi.BrightBlack(cmd.Desc))
		sb.WriteRune('\n')
	}
	return sb.String()
}
