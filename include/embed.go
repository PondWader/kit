package include

import (
	"embed"
	_ "embed"
)

//go:embed repositories.kit
var Repositories string

//go:embed migrations/*
var Migrations embed.FS

//go:embed lib/*
var Lib embed.FS
