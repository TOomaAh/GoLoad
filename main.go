package main

import (
	"embed"

	"github.com/TOomaAh/GoLoad/cmd"
)

//go:embed frontend/dist
var assets embed.FS

func main() {
	cmd.Run(&assets)
}
