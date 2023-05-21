package main

import (
	"embed"
	"io/fs"
	"log"

	"github.com/xbugio/go-vhostd/internal/cmd"
)

//go:embed html
var htmlFs embed.FS

func main() {
	subFs, err := fs.Sub(&htmlFs, "html")
	if err != nil {
		log.Fatal(err)
	}
	cmd.NewCmd(subFs).Execute()
}
