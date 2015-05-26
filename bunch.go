package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/codegangsta/cli"
	"github.com/kardianos/osext"
)

var InitialPath string
var InitialGoPath string

var Verbose bool

var SpinnerCharSet = 14
var SpinnerInterval = 50 * time.Millisecond

func main() {
	currentExecutable, _ := osext.Executable()
	vendoredBunchPath := path.Join(".vendor", "bin", "bunch")

	fi1, errStat1 := os.Stat(currentExecutable)
	fi2, errStat2 := os.Stat(vendoredBunchPath)

	if exists, _ := pathExists(vendoredBunchPath); errStat1 == nil && errStat2 == nil && exists && !os.SameFile(fi1, fi2) {
		cmd := exec.Command(vendoredBunchPath, os.Args[1:]...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err := cmd.Run()
		if err == nil {
			if cmd.ProcessState.Success() {
				os.Exit(0)
			}
		}

		// if "subbunch" succeeded, exit, otherwise, continue with regular bunch business
		fmt.Println("vendored bunch exited with a non-zero exit status, trying again with global bunch")
	}

	InitialPath = os.Getenv("PATH")
	InitialGoPath = os.Getenv("GOPATH")

	app := cli.NewApp()
	app.Name = "bunch"
	app.Usage = "npm-like tool for managing Go dependencies"
	app.Version = "0.0.1"
	app.Authors = []cli.Author{cli.Author{Name: "Daniil Kulchenko", Email: "daniil@kulchenko.com"}}

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "verbose",
			Usage: "output more information",
		},
	}

	app.Before = func(context *cli.Context) error {
		Verbose = context.GlobalBool("verbose")

		return nil
	}

	app.Commands = []cli.Command{
		{
			Name:    "install",
			Aliases: []string{"i"},
			Usage:   "install package(s)",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "save",
					Usage: "save installed package to Bunchfile",
				},
				cli.BoolFlag{
					Name:  "g",
					Usage: "install package to global $GOPATH instead of vendored directory",
				},
			},
			Action: func(c *cli.Context) {
				installCommand(c, false, true, true)
			},
		},
		{
			Name:    "update",
			Aliases: []string{"u"},
			Usage:   "update package(s)",
			Action: func(c *cli.Context) {
				installCommand(c, true, true, false)
			},
		},
		{
			Name:    "uninstall",
			Aliases: []string{"r"},
			Usage:   "uninstall package(s)",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "save",
					Usage: "save uninstalled package to Bunchfile",
				},
				cli.BoolFlag{
					Name:  "g",
					Usage: "uninstall package from global $GOPATH instead of vendored directory",
				},
			},
			Action: func(c *cli.Context) {
				uninstallCommand(c)
			},
		},
		{
			Name:  "prune",
			Usage: "remove packages not referenced in Bunchfile",
			Action: func(c *cli.Context) {
				pruneCommand(c)
			},
		},
		{
			Name:  "outdated",
			Usage: "list outdated packages",
			Action: func(c *cli.Context) {
				outdatedCommand(c)
			},
		},
		{
			Name:  "lock",
			Usage: "generate a file locking down current versions of dependencies",
			Action: func(c *cli.Context) {
				lockCommand(c)
			},
		},
		{
			Name:  "rebuild",
			Usage: "rebuild all dependencies",
			Action: func(c *cli.Context) {
				installCommand(c, true, false, true)
			},
		},
		{
			Name:  "generate",
			Usage: "generate a Bunchfile based on package imports in current directory",
			Action: func(c *cli.Context) {
				generateCommand(c)
			},
		},
		{
			Name:            "go",
			Usage:           "run a Go command within the vendor environment (e.g. bunch go fmt)",
			SkipFlagParsing: true,
			Action: func(c *cli.Context) {
				goCommand(c)
			},
		},
		{
			Name:            "exec",
			Usage:           "run any command within the vendor environment (e.g. bunch exec make)",
			SkipFlagParsing: true,
			Action: func(c *cli.Context) {
				execCommand(c)
			},
		},
		{
			Name:            "shell",
			Usage:           "start a shell within the vendor environment",
			SkipFlagParsing: true,
			Action: func(c *cli.Context) {
				shellCommand(c)
			},
		},
		{
			Name:  "shim",
			Usage: "sourced in .bash_profile to alias the 'go' tool",
			Action: func(c *cli.Context) {
				shimCommand(c)
			},
		},
	}

	app.Run(os.Args)
}
