package main

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	version "github.com/hashicorp/go-version"
	"github.com/juju/errors"
)

func getLatestVersionMatchingPattern(repo string, versionPattern string) (string, error) {
	repoPath, err := getPackageRootDir(repo)
	if err != nil {
		return "", errors.Trace(err)
	}

	if exists, _ := pathExists(repoPath); !exists {
		return versionPattern, nil
	}

	wd, err := os.Getwd()
	if err != nil {
		return "", errors.Trace(err)
	}

	err = os.Chdir(repoPath)
	if err != nil {
		return "", errors.Trace(err)
	}

	defer func() {
		_ = os.Chdir(wd)
	}()

	if exists, _ := pathExists(".git"); !exists {
		return versionPattern, nil // for now, we only know git
	} else {
		if versionPattern == "" {
			return "master", nil
		}
	}

	// first, try feeding it through git to see if it's a valid rev
	gitResolveCommand := []string{"git", "rev-parse", "-q", "--verify", versionPattern}
	output, err := exec.Command(gitResolveCommand[0], gitResolveCommand[1:]...).Output()

	if err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			return "", errors.Trace(err)
		}
	}

	gitResolvedString := strings.TrimSpace(string(output))

	if gitResolvedString != "" {
		return gitResolvedString, nil
	}

	// second, try parsing it
	tagListB, err := exec.Command("git", "tag").Output()
	if err != nil {
		return "", errors.Trace(err)
	}

	versionToTag := make(map[*version.Version]string)

	tagList := strings.Split(strings.TrimSpace(string(tagListB)), "\n")
	processedTagList := make([]*version.Version, len(tagList))
	for i, tag := range tagList {
		stringVersion := tag

		if strings.HasPrefix(tag, "v") {
			stringVersion = strings.Replace(tag, "v", "", 1)
		}

		v, err := version.NewVersion(stringVersion)
		if err != nil {
			continue
		}

		processedTagList[i] = v
		versionToTag[v] = tag
	}

	sort.Sort(version.Collection(processedTagList))

	constraints, err := version.NewConstraint(versionPattern)
	if err != nil {
		return "", errors.Trace(err)
	}

	var resultVersion string
	for i := len(processedTagList) - 1; i >= 0; i-- {
		ver := processedTagList[i]
		if constraints.Check(ver) {
			resultVersion = versionToTag[ver]
		}
	}

	if resultVersion == "" {
		return "", fmt.Errorf("unable to find a version matching constraint %s for package %s", versionPattern, repo)
	}

	gitResolveCommand = []string{"git", "rev-parse", "-q", "--verify", resultVersion}
	output, err = exec.Command(gitResolveCommand[0], gitResolveCommand[1:]...).Output()

	if err != nil {
		return "", errors.Trace(err)
	} else {
		return strings.TrimSpace(string(output)), nil
	}
}
