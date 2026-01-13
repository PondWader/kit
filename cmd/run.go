package main

import (
	"bytes"
	"fmt"
	"log"

	"github.com/PondWader/kit/pkg/lang"
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

	code := `export name = "go"

export fn install(version) {
    resp = fetch("https://go.dev/dl/go${version}.${sys.OS}-${sys.ARCH}.tar.gz")
    tar.gz.extract(resp).to("/")
    link_bin_dir("/bin")
}

export fn versions() {
    return fetch("https://proxy.golang.org/golang.org/toolchain/@v/list")
        .text()
        .trim_whitespace()
        .split("\n")
        .map_to_set(l -> 
            l.cut_prefix_before("-").cut_suffix_after(".")
        )
}
`
	prog, err := lang.Parse(bytes.NewReader([]byte(code)))
	if err != nil {
		log.Fatalln(err)
	}

	env := lang.NewEnv()
	env.LoadStd()
	if err := env.Execute(prog); err != nil {
		log.Fatalln(err)
	}

	fmt.Println(env.Exports["versions"].Call())

}
