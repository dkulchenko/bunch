package main

import (
	"encoding/json"
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/juju/errors"
)

type Package struct {
	Repo          string
	Version       string
	LockedVersion string

	IsSelf     bool
	IsLink     bool
	LinkTarget string
}

type BunchFile struct {
	Packages []Package
	Raw      []string
}

var commentStripRegexp = regexp.MustCompile(`#.*`)
var versionSwapRegexp = regexp.MustCompile(`^(\S+)\s*(.*)`)

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
		raw := []string{pack.Repo}
		if pack.Version != "" {
			raw = append(raw, pack.Version)
		}

		b.Raw = append(b.Raw, strings.Join(raw, " "))
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
	err := ioutil.WriteFile("Bunchfile", []byte(strings.Join(append(b.Raw, ""), "\n")), 0644)

	if err != nil {
		return errors.Trace(err)
	}

	return nil
}

func createBunchfile() *BunchFile {
	return &BunchFile{}
}

func readBunchfile() (*BunchFile, error) {
	bunchbytes, err := ioutil.ReadFile("Bunchfile")

	if err != nil {
		return &BunchFile{}, errors.Trace(err)
	}

	bunch := BunchFile{
		Raw: strings.Split(strings.TrimSpace(string(bunchbytes)), "\n"),
	}

	lockedCommits := make(map[string]string)

	if exists, _ := pathExists("Bunchfile.lock"); exists {
		lockBytes, err := ioutil.ReadFile("Bunchfile.lock")
		if err != nil {
			return &BunchFile{}, errors.Trace(err)
		}

		err = json.Unmarshal(lockBytes, &lockedCommits)
		if err != nil {
			return &BunchFile{}, errors.Trace(err)
		}
	}

	for _, line := range bunch.Raw {
		line = commentStripRegexp.ReplaceAllLiteralString(line, "")
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		pack := Package{}

		packageInfo := strings.SplitN(line, " ", 2)

		if len(packageInfo) < 1 {
			continue
		} else {
			pack.Repo = packageInfo[0]
		}

		if len(packageInfo) >= 2 {
			pack.Version = packageInfo[1]
		}

		if strings.HasPrefix(pack.Version, "!link") || strings.HasPrefix(pack.Version, "!self") {
			if strings.HasPrefix(pack.Version, "!self") {
				pack.IsSelf = true
			}

			pack.IsLink = true

			linkList := strings.Split(pack.Version, ":")

			if len(linkList) == 2 {
				pack.LinkTarget = linkList[1]
			} else {
				wd, err := os.Getwd()
				if err != nil {
					return &BunchFile{}, errors.Trace(err)
				}

				pack.LinkTarget = wd
			}
		}

		if lockedCommits[pack.Repo] != "" {
			pack.LockedVersion = lockedCommits[pack.Repo]
		}

		bunch.Packages = append(bunch.Packages, pack)
	}

	return &bunch, nil
}

func filterCommonBasePackages(depList []string, selfBase string) []string {
	basePackages := []string{}

	for i, dep1 := range depList {
		foundPrefix := false

		if strings.HasPrefix(dep1, selfBase) {
			continue
		}

		for j, dep2 := range depList {
			if i == j {
				continue
			}

			if strings.HasPrefix(dep1, dep2) {
				foundPrefix = true
				break
			}
		}

		if !foundPrefix {
			basePackages = append(basePackages, dep1)
		}
	}

	return basePackages
}

func generateBunchfile() error {
	bunch := BunchFile{}

	goListCommand := []string{"go", "list", "--json", "."}
	output, err := exec.Command(goListCommand[0], goListCommand[1:]...).Output()
	if err != nil {
		return errors.Trace(err)
	}

	packageInfo := GoList{}
	err = json.Unmarshal(output, &packageInfo)

	if err != nil {
		return errors.Trace(err)
	}

	err = bunch.AddPackage(fmt.Sprintf("%s@!self", packageInfo.ImportPath))
	if err != nil {
		return errors.Trace(err)
	}

	for _, dep := range filterCommonBasePackages(append(packageInfo.Deps, packageInfo.TestImports...), packageInfo.ImportPath) {
		// check that the package is not part of the standard library
		if exists, _ := pathExists(path.Join(build.Default.GOROOT, "src", dep)); !exists {
			err = bunch.AddPackage(dep)
			if err != nil {
				return errors.Trace(err)
			}
		}
	}

	err = bunch.Save()
	if err != nil {
		return errors.Trace(err)
	}

	color.Green("Bunchfile generated successfully")

	return nil
}
