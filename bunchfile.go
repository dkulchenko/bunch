package main

import (
	"io/ioutil"
	"regexp"
	"strings"
)

type Package struct {
	Repo    string
	Version string
}

type BunchFile struct {
	Packages    []Package
	DevPackages []Package
	Raw         []string
}

var commentStripRegexp = regexp.MustCompile(`#.*`)

func (b *BunchFile) AddPackage(pack string) error {
	return nil
}

func (b *BunchFile) Save() error {
	return nil
}

func readBunchfile() (*BunchFile, error) {
	bunchbytes, err := ioutil.ReadFile("Bunchfile")

	if err != nil {
		return &BunchFile{}, err
	}

	bunch := BunchFile{
		Raw: strings.Split(string(bunchbytes), "\n"),
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

		bunch.Packages = append(bunch.Packages, Package{Repo: repo, Version: version})
	}

	return &bunch, nil
}
