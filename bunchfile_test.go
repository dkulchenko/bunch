package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterCommonBasePackages(t *testing.T) {
	depList := []string{"github.com/a/b/c", "github.com/d/e/f", "github.com/a/b"}
	result := filterCommonBasePackages(depList, "github.com/my/package")

	expectedResult := []string{"github.com/d/e/f", "github.com/a/b"}

	assert.Equal(t, result, expectedResult, "packages with a common base should be filtered out")
}

func TestCreateBunchfile(t *testing.T) {
	assert.Equal(t, &BunchFile{}, createBunchfile(), "create bunchfile returns a bunchfile object")
}

func TestRawIndex(t *testing.T) {
	bunch := createBunchfile()

	bunch.Raw = []string{"github.com/a/b", "github.com/a/b/c", "github.com/d 123456"}

	rawIndex, present := bunch.RawIndex("github.com/a/b")
	assert.Equal(t, rawIndex, 0, "found github.com/a/b at first position in bunchfile")
	assert.Equal(t, present, true, "found github.com/a/b at first position in bunchfile")

	rawIndex, present = bunch.RawIndex("github.com/d")
	assert.Equal(t, rawIndex, 2, "found github.com/d at third position in bunchfile")
	assert.Equal(t, present, true, "found github.com/a/b at first position in bunchfile")

	rawIndex, present = bunch.RawIndex("github.com/a/b/c")
	assert.Equal(t, rawIndex, 1, "found github.com/a/b/c at second position in bunchfile")
	assert.Equal(t, present, true, "found github.com/a/b/c in bunchfile")

	_, present = bunch.RawIndex("github.com/unknown")
	assert.Equal(t, present, false, "did not find github.com/unknown in bunchfile")
}

func TestPackageIndex(t *testing.T) {
	bunch := createBunchfile()

	bunch.Packages = []Package{
		Package{
			Repo:    "github.com/a/b",
			Version: "123",
		},
		Package{
			Repo: "github.com/def",
		},
		Package{
			Repo:    "github.com/a/b/c",
			Version: "xyz",
		},
	}

	packageIndex, present := bunch.PackageIndex("github.com/a/b")
	assert.Equal(t, packageIndex, 0, "found github.com/a/b in first position in bunchfile package list")
	assert.Equal(t, present, true, "found github.com/a/b in bunchfile package list")

	packageIndex, present = bunch.PackageIndex("github.com/def")
	assert.Equal(t, packageIndex, 1, "found github.com/def at second position in bunchfile package list")
	assert.Equal(t, present, true, "found github.com/def in bunchfile package list")

	packageIndex, present = bunch.PackageIndex("github.com/a/b/c")
	assert.Equal(t, packageIndex, 2, "found github.com/a/b/c at third position in bunchfile package list")
	assert.Equal(t, present, true, "found github.com/a/b/c in bunchfile package list")

	_, present = bunch.PackageIndex("github.com/unknown")
	assert.Equal(t, present, false, "did not find github.com/unknown in bunchfile package list")
}

func TestAddPackageNew(t *testing.T) {
	bunch := createBunchfile()

	err := bunch.AddPackage("github.com/a/b/c")
	assert.Nil(t, err, "did not error on adding package")

	packageIndex, present := bunch.PackageIndex("github.com/a/b/c")
	assert.Equal(t, packageIndex, 0, "found github.com/a/b/c in first position in bunchfile package list")
	assert.Equal(t, present, true, "found github.com/a/b/c in bunchfile package list")

	err = bunch.AddPackage("github.com/a/b/d@v1.2.0")
	assert.Nil(t, err, "did not error on adding package")

	packageIndex, present = bunch.PackageIndex("github.com/a/b/d")
	assert.Equal(t, packageIndex, 1, "found github.com/a/b/d in second position in bunchfile package list")
	assert.Equal(t, present, true, "found github.com/a/b/d in bunchfile package list")
}

func TestAddPackageUpdate(t *testing.T) {
	bunch := createBunchfile()

	err := bunch.AddPackage("github.com/a/b/c")
	assert.Nil(t, err, "did not error on adding package")

	pack := bunch.Packages[0]
	assert.Equal(t, len(bunch.Packages), 1, "should only be one package in bunchfile")
	assert.Equal(t, "", pack.Version, "version should not be set yet")

	err = bunch.AddPackage("github.com/a/b/c@v1.2.1")
	assert.Nil(t, err, "did not error on adding package")

	pack = bunch.Packages[0]
	assert.Equal(t, len(bunch.Packages), 1, "should still be only one package in bunchfile")
	assert.Equal(t, pack.Version, "v1.2.1", "version should now be set")
}

func TestRemovePackage(t *testing.T) {
	bunch := createBunchfile()

	err := bunch.AddPackage("github.com/a/b/c")
	assert.Nil(t, err, "did not error on adding package")

	assert.Equal(t, len(bunch.Raw), 1, "should be one packages in Bunchfile")
	assert.Equal(t, len(bunch.Packages), 1, "should be one package in bunchfile package list")

	err = bunch.RemovePackage("github.com/a/b/c")
	assert.Nil(t, err, "did not error on removing package")

	assert.Equal(t, len(bunch.Raw), 0, "should be no packages in Bunchfile")
	assert.Equal(t, len(bunch.Packages), 0, "should be no packages in Bunchfile package list")
}
