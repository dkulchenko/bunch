package main

import (
	"os"

	"github.com/codegangsta/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "bunch"
	app.Usage = "npm-like tool for managing Go dependencies"
	app.Version = "0.0.1"
	app.Authors = []cli.Author{cli.Author{Name: "Daniil Kulchenko", Email: "daniil@kulchenko.com"}}

	app.Run(os.Args)
}
