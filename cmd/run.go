package main

import (
	"bytes"
	"fmt"
	"log"

	"github.com/PondWader/kit/pkg/lang"
)

func main() {
	code := `export name = "go"

export fn install(version) {
    resp = fetch("https://go.dev/dl/go${version}.${sys.OS}-${sys.ARCH}.tar.gz")
    tar.gz.extract(resp).to("/")
    link_bin_dir("/bin")
}

export fn versions() {
    return fetch("https://proxy.golang.org/golang.org/toolchain/@v/list")
        .text()
        .trim()
        .split("\n")
        .map_to_set(l -> 
            l.cut_prefix_before('-').cut_suffix_after('.')
        )
}
`
	prog, err := lang.Parse(bytes.NewReader([]byte(code)))
	if err != nil {
		log.Fatalln(err)
	}

	env := lang.NewEnv()
	if err := env.Execute(prog); err != nil {
		log.Fatalln(err)
	}

	fmt.Println(
		env.Exports,
		env.Vars,
	)
}
