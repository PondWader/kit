package main

import (
	"fmt"
	"strings"

	"github.com/PondWader/kit/internal/ansi"
)

func main() {
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
		"            ")

	fmt.Println()
	fmt.Println(fmtCommandMenu([]cmd{
		{Args: "install <package>[@version] (alias: add)", Desc: "install a package"},
		{Args: "uninstall <package>[@version] (alias: remove)", Desc: "uninstall a package"},
		{Args: "use <package>@<version>", Desc: "switch to a specific version of a package"},
		{Args: "list [repos/packages/available] (alias: ls)", Desc: "lists all repositories, installed packages or available packages (default: installed packages)"},
		{Args: "versions <package>", Desc: "lists all versions available for a package"},
		{Args: "search <term>", Desc: "search packages"},
		{Args: "pull", Desc: "pulls the latest version of all repositories"},
	}))
	fmt.Println()

	// 	code := `export name = "go"

	// export fn install(version) {
	//     resp = fetch("https://go.dev/dl/go${version}.${sys.OS}-${sys.ARCH}.tar.gz")
	//     tar.gz.extract(resp).to("/")
	//     link_bin_dir("/bin")
	// }

	// export fn versions() {
	//     return fetch("https://proxy.golang.org/golang.org/toolchain/@v/list")
	//         .text()
	//         .trim_whitespace()
	//         .split("\n")
	//         .map(l ->
	//             l.cut_prefix_before("-").cut_suffix_after(".")
	//         )
	// }
	// `
	// prog, err := lang.Parse(bytes.NewReader([]byte(code)))
	// if err != nil {
	// 	log.Fatalln(err)
	// }

	// env := lang.NewEnv()
	// env.LoadStd()
	// if err := env.Execute(prog); err != nil {
	// 	log.Fatalln(err)
	// }

	// v, e := env.Exports["versions"].Call()
	// if e != nil {
	// 	log.Fatalln(e)
	// }
	// l, ok := v.ToList()
	// if !ok {
	// 	log.Fatalln("expected list but got ", v.Kind())
	// }
	// s := l.AsSlice()
	// // for _, ver := range s {
	// // 	// fmt.Println(ver)
	// // }
	// _ = s

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
