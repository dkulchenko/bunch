package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"

	"github.com/fatih/color"
)

type Package struct {
	Repo          string
	Version       string
	LockedVersion string
}

type BunchFile struct {
	Packages []Package
	Raw      []string
}

var commentStripRegexp = regexp.MustCompile(`#.*`)
var versionSwapRegexp = regexp.MustCompile(`^(\S+)\s*(\S*)`)

func (b *BunchFile) RawIndex(repo string) (int, bool) {
	for i, packString := range b.Raw {
		parts := strings.Fields(packString)

		if len(parts) < 1 {
			continue
		}

		if parts[0] == repo {
			return i, true
		}
	}

	return 0, false
}

func (b *BunchFile) PackageIndex(repo string) (int, bool) {
	for i, pack := range b.Packages {
		if pack.Repo == repo {
			return i, true
		}
	}

	return 0, false
}

func (b *BunchFile) AddPackage(packString string) error {
	pack := parsePackage(packString)

	index, present := b.RawIndex(pack.Repo)

	if present {
		packIndex, _ := b.PackageIndex(pack.Repo)
		b.Packages[packIndex] = pack

		initialLine := b.Raw[index]

		replacementString := fmt.Sprintf("$1 %s", pack.Version)
		newLine := versionSwapRegexp.ReplaceAllString(initialLine, replacementString)

		b.Raw[index] = newLine
	} else {
		b.Packages = append(b.Packages, pack)
		b.Raw = append(b.Raw, fmt.Sprintf("%s %s", pack.Repo, pack.Version))
	}

	return nil
}

func (b *BunchFile) RemovePackage(packString string) error {
	pack := parsePackage(packString)

	index, present := b.RawIndex(pack.Repo)

	if present {
		packIndex, packPresent := b.PackageIndex(pack.Repo)

		if packPresent {
			if packIndex < len(b.Packages)-1 {
				b.Packages = append(b.Packages[:packIndex], b.Packages[packIndex+1:]...)
			} else {
				b.Packages = b.Packages[:packIndex]
			}
		}

		if index < len(b.Raw)-1 {
			b.Raw = append(b.Raw[:index], b.Raw[index+1:]...)
		} else {
			b.Raw = b.Raw[:index]
		}
	}

	return nil
}

func (b *BunchFile) Save() error {
	err := ioutil.WriteFile("Bunchfile", []byte(strings.Join(b.Raw, "\n")), 0644)

	if err != nil {
		return err
	}

	return nil
}

func createBunchfile() *BunchFile {
	return &BunchFile{}
}

func readBunchfile() (*BunchFile, error) {
	bunchbytes, err := ioutil.ReadFile("Bunchfile")

	if err != nil {
		return &BunchFile{}, err
	}

	bunch := BunchFile{
		Raw: strings.Split(strings.TrimSpace(string(bunchbytes)), "\n"),
	}

	lockedCommits := make(map[string]string)

	if exists, _ := pathExists("Bunchfile.lock"); exists {
		lockBytes, err := ioutil.ReadFile("Bunchfile.lock")
		if err != nil {
			return &BunchFile{}, err
		}

		err = json.Unmarshal(lockBytes, &lockedCommits)
		if err != nil {
			return &BunchFile{}, err
		}
	}

	for _, line := range bunch.Raw {
		line = commentStripRegexp.ReplaceAllLiteralString(line, "")
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		var repo, version string

		packageInfo := strings.Fields(line)

		if len(packageInfo) < 1 {
			continue
		} else {
			repo = packageInfo[0]
		}

		if len(packageInfo) >= 2 {
			version = packageInfo[1]
		}

		var lockedVersion string
		if lockedCommits[repo] != "" {
			lockedVersion = lockedCommits[repo]
		}

		bunch.Packages = append(bunch.Packages, Package{Repo: repo, Version: version, LockedVersion: lockedVersion})
	}

	return &bunch, nil
}

func generateBunchfile() error {
	bunch := BunchFile{}

	gopath := os.Getenv("GOPATH")

	goListCommand := []string{"go", "list", "--json", "."}
	output, err := exec.Command(goListCommand[0], goListCommand[1:]...).Output()
	if err != nil {
		return err
	}

	packageInfo := GoList{}
	err = json.Unmarshal(output, &packageInfo)

	if err != nil {
		return err
	}

	for _, dep := range packageInfo.Deps {
		depPath := path.Join(gopath, "src", dep)
		if exists, _ := pathExists(depPath); exists {
			err = bunch.AddPackage(dep)
			if err != nil {
				return err
			}
		}
	}

	err = bunch.Save()
	if err != nil {
		return err
	}

	color.Green("Bunchfile generated successfully")

	return nil
}
