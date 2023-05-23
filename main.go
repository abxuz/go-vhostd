package main

import (
	"embed"

	"github.com/xbugio/go-vhostd/internal/cmd"
)

//go:embed html
var htmlFs embed.FS

func main() {
	cmd.NewCmd(&htmlFs).Execute()
}
