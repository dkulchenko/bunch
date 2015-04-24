package main

import (
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/codegangsta/cli"
)

func setupVendoring() error {
	vendorDirs := []string{".vendor/bin", ".vendor/pkg", ".vendor/src"}

	for _, vendorDir := range vendorDirs {
		err := os.MkdirAll(vendorDir, 0755)

		if err != nil {
			return err
		}
	}

	return nil
}

func installCommand(c *cli.Context) {
	// bunch install
	// bunch install github.com/abc/xyz
	// bunch install github.com/abc/xyz github.com/abc/def
	// bunch install github.com/abc/xyz --save
	// bunch install github.com/abc/xyz -g
	// bunch install abc/xyz # github shorthand

	packages := c.Args()

	err := setupVendoring()
	if err != nil {
		log.Fatalf("unable to set up vendor dirs: %s", err)
	}

	if len(packages) == 0 {
		bunch, err := readBunchfile()
		if err != nil {
			log.Fatalf("unable to read Bunchfile: %s", err)
		}

		err = installPackagesFromBunchfile(bunch)

		if err != nil {
			log.Fatalf("failed installing packages: %s", err)
		}
	} else {
		global := c.Bool("g")
		save := c.Bool("save")

		var bunch *BunchFile
		if exists, _ := pathExists("Bunchfile"); exists {
			bunch, err = readBunchfile()
			if err != nil {
				log.Fatalf("unable to read Bunchfile: %s", err)
			}
		} else {
			bunch = createBunchfile()
		}

		err := installPackages(packages, global)
		if err != nil {
			log.Fatalf("failed installing packages: %s", err)
		}

		if save {
			for _, pack := range packages {
				err := bunch.AddPackage(pack)

				if err != nil {
					log.Fatalf("failed adding package %s to save list: %s", pack, err)
				}
			}

			err = bunch.Save()
			if err != nil {
				log.Fatalf("failed saving Bunchfile: %s", err)
			}
		}
	}
}

func updateCommand(c *cli.Context) {
	// bunch update
	// bunch update github.com/abc/xyz
	// bunch update github.com/abc/xyz github.com/abc/def
	// bunch update github.com/abc/xyz --save
	// bunch update github.com/abc/xyz -g
}

func uninstallCommand(c *cli.Context) {
	// bunch uninstall github.com/abc/xyz
	// bunch uninstall github.com/abc/xyz --save
	// bunch uninstall github.com/abc/xyz -g

	// use go list --json to remove unreferences deps (when not using -g)
}

func pruneCommand(c *cli.Context) {
	// bunch prune

	// use go list --json
}

func outdatedCommand(c *cli.Context) {
	// bunch outdated
}

func lockCommand(c *cli.Context) {
	// bunch lock
}

func rebuildCommand(c *cli.Context) {
	// bunch rebuild
}

func generateCommand(c *cli.Context) {
	// bunch generate

	// use go list --json
}

func goCommand(c *cli.Context) {
	// bunch go test
	// bunch go fmt
	// bunch go ...

	err := setVendorEnv()
	if err != nil {
		log.Fatalf("unable to set vendor env: %s", err)
	}

	cmd := exec.Command("go", c.Args()...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		log.Fatalf("running 'go %s' failed: %s", strings.Join(c.Args(), " "), err)
	}
}

func execCommand(c *cli.Context) {
	// bunch exec make
}

func shellCommand(c *cli.Context) {
	// bunch shell (bunch exec $SHELL)
}

func shimCommand(c *cli.Context) {
	// bunch shim outputs help text
	// bunch shim - outputs a shell script
	// in .bash_profile...
	// if which bunch > /dev/null; then eval "$(bunch shim -)"; fi
}
