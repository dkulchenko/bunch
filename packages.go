package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
)

func parseRepo(repo string) string {
	if strings.HasPrefix(repo, "github.com") {
		return fmt.Sprintf("https://%s", repo)
	} else {
		return repo
	}
}

func fetchPackage(repo string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	packageDir := path.Join(wd, ".vendor", "src", repo)

	if _, err := os.Stat(packageDir); err != nil {
		if os.IsNotExist(err) {
			s := spinner.New(spinner.CharSets[9], 50*time.Millisecond)
			s.Prefix = fmt.Sprintf("fetching %s ", repo)
			s.Color("green")
			s.Start()

			gitRepo := parseRepo(repo)

			gitCloneCommand := []string{"git", "clone", gitRepo, packageDir}
			gitCloneOutput, err := exec.Command(gitCloneCommand[0], gitCloneCommand[1:]...).CombinedOutput()

			s.Stop()
			fmt.Printf("\rfetching %s ... %s\n", repo, color.GreenString("done"))

			if err != nil {
				return errors.New(fmt.Sprintf("failed cloning git repo for package %s, error: %s, output: %s", repo, err, gitCloneOutput))
			}

			return nil
		}
	}

	err = os.Chdir(packageDir)
	if err != nil {
		return err
	}

	s := spinner.New(spinner.CharSets[9], 50*time.Millisecond)
	s.Prefix = fmt.Sprintf("refreshing %s ", repo)
	s.Color("green")
	s.Start()

	gitFetchCommand := []string{"git", "fetch", "--all"}
	gitFetchOutput, err := exec.Command(gitFetchCommand[0], gitFetchCommand[1:]...).CombinedOutput()

	s.Stop()
	fmt.Printf("\rrefreshing %s ... %s\n", repo, color.GreenString("done"))

	if err != nil {
		return errors.New(fmt.Sprintf("failed updating git repo for package %s, error: %s, output: %s", repo, err, gitFetchOutput))
	}

	err = os.Chdir(wd)
	if err != nil {
		return err
	}

	return nil
}

func fetchPackageDependencies(repo string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	packageDir := path.Join(wd, ".vendor", "src", repo)

	err = os.Chdir(packageDir)
	if err != nil {
		return err
	}

	s := spinner.New(spinner.CharSets[9], 50*time.Millisecond)
	s.Prefix = fmt.Sprintf("  - fetching dependencies for %s ", repo)
	s.Color("green")
	s.Start()

	goGetCommand := []string{"go", "get", "-u", "-d", "./..."}
	goGetOutput, err := exec.Command(goGetCommand[0], goGetCommand[1:]...).CombinedOutput()

	s.Stop()
	fmt.Printf("\r  - fetching dependencies for %s ... %s\n", repo, color.GreenString("done"))

	if err != nil {
		return errors.New(fmt.Sprintf("failed fetching dependencies forpackage %s, error: %s, output: %s", repo, err, goGetOutput))
	}

	err = os.Chdir(wd)
	if err != nil {
		return err
	}

	return nil
}

func buildPackage(repo string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	packageDir := path.Join(wd, ".vendor", "src", repo)

	err = os.Chdir(packageDir)
	if err != nil {
		return err
	}

	s := spinner.New(spinner.CharSets[9], 50*time.Millisecond)
	s.Prefix = fmt.Sprintf("  - building package %s ", repo)
	s.Color("green")
	s.Start()

	goBuildCommand := []string{"go", "build", repo}
	goBuildOutput, err := exec.Command(goBuildCommand[0], goBuildCommand[1:]...).CombinedOutput()

	s.Stop()
	fmt.Printf("\r  - building package %s ... %s\n", repo, color.GreenString("done"))

	if err != nil {
		return errors.New(fmt.Sprintf("failed building package %s, error: %s, output: %s", repo, err, goBuildOutput))
	}

	err = os.Chdir(wd)
	if err != nil {
		return err
	}

	return nil
}

func installPackage(repo string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	packageDir := path.Join(wd, ".vendor", "src", repo)

	err = os.Chdir(packageDir)
	if err != nil {
		return err
	}

	s := spinner.New(spinner.CharSets[9], 50*time.Millisecond)
	s.Prefix = fmt.Sprintf("  - installing package %s ", repo)
	s.Color("green")
	s.Start()

	goInstallCommand := []string{"go", "install", repo}
	goInstallOutput, err := exec.Command(goInstallCommand[0], goInstallCommand[1:]...).CombinedOutput()

	s.Stop()
	fmt.Printf("\r  - installing package %s ... %s\n", repo, color.GreenString("done"))

	if err != nil {
		return errors.New(fmt.Sprintf("failed installing package %s, error: %s, output: %s", repo, err, goInstallOutput))
	}

	err = os.Chdir(wd)
	if err != nil {
		return err
	}

	return nil
}

func setPackageVersion(repo string, version string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	packageDir := path.Join(wd, ".vendor", "src", repo)

	err = os.Chdir(packageDir)
	if err != nil {
		return err
	}

	s := spinner.New(spinner.CharSets[9], 50*time.Millisecond)
	s.Prefix = fmt.Sprintf("  - setting version of %s to %s ", repo, version)
	s.Color("green")
	s.Start()

	checkoutCommand := []string{"git", "checkout", version}
	checkoutOutput, err := exec.Command(checkoutCommand[0], checkoutCommand[1:]...).CombinedOutput()

	s.Stop()
	fmt.Printf("\r  - setting version of %s to %s ... %s\n", repo, version, color.GreenString("done"))

	if err != nil {
		return errors.New(fmt.Sprintf("failed setting version of package %s, error: %s, output: %s", repo, err, checkoutOutput))
	}

	err = os.Chdir(wd)
	if err != nil {
		return err
	}

	return nil
}

func installPackagesFromBunchfile(b *BunchFile) error {
	err := setVendorEnv()
	if err != nil {
		return err
	}

	for _, pack := range b.Packages {
		err = fetchPackage(pack.Repo)
		if err != nil {
			return err
		}

		err = fetchPackageDependencies(pack.Repo)
		if err != nil {
			return err
		}

		err = setPackageVersion(pack.Repo, pack.Version)
		if err != nil {
			return err
		}

		err = buildPackage(pack.Repo)
		if err != nil {
			return err
		}

		err = installPackage(pack.Repo)
		if err != nil {
			return err
		}

		fmt.Print(color.GreenString("\rsuccessfully installed %s                 \n\n", pack.Repo))
	}

	return nil
}

func installPackages(packages []string, installGlobally bool) error {
	return nil
}
