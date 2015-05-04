package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPathExists(t *testing.T) {
	exists, err := pathExists("/")
	assert.Nil(t, err, "path exists should not error out")

	assert.Equal(t, exists, true, "/ path should exist")
}

func TestCountNonEmptyString(t *testing.T) {
	emptiesGalore := []string{"a", "b", "", "c", "", "", "d", ""}
	noEmpties := []string{"a", "b", "c"}

	assert.Equal(t, 4, countNonEmptyStrings(emptiesGalore), "should count 4 non-empties")
	assert.Equal(t, 3, countNonEmptyStrings(noEmpties), "should count 3 non-empties")
}

func TestParsePackage(t *testing.T) {
	pack1 := parsePackage("github.com/a/b/c")
	pack2 := parsePackage("github.com/a/b/c@v1.2.0")
	pack3 := parsePackage("a/b")
	pack4 := parsePackage("gopkg.in/abc")

	assert.Equal(t, pack1.Repo, "github.com/a/b/c", "package repo should equal")
	assert.Equal(t, pack2.Repo, "github.com/a/b/c", "package repo should equal")

	assert.Equal(t, pack1.Version, "", "package version should be unset")
	assert.Equal(t, pack2.Version, "v1.2.0", "package version should be set")

	assert.Equal(t, pack3.Repo, "github.com/a/b", "github shorthand should have been expanded")
	assert.Equal(t, pack4.Repo, "gopkg.in/abc", "package containing domain should not have been expanded")
}

/*
  pack1 := parsePackage("github.com/a/b/c")
  pack2 := parsePackage("github.com/a/b/c !self")
  pack3 := parsePackage("github.com/a/b/c !link:/tmp")
  pack4 := parsePackage("github.com/a/b/c v1.2.0")
  pack5 := parsePackage("github.com/a/b/c    master")

  assert.Equal(t, pack1.Repo, "github.com/a/b/c", "package repo should equal")
  assert.Equal(t, pack2.Repo, "github.com/a/b/c", "package repo should equal")
  assert.Equal(t, pack3.Repo, "github.com/a/b/c", "package repo should equal")
  assert.Equal(t, pack4.Repo, "github.com/a/b/c", "package repo should equal")
  assert.Equal(t, pack5.Repo, "github.com/a/b/c", "package repo should equal")

  assert.Equal(t, pack1.Version, "", "package version should be unset")
  assert.Equal(t, pack4.Version, "v1.2.0", "package version should be set")
  assert.Equal(t, pack5.Version, "master", "package version should be set")

  assert.Equal(t, pack2.IsSelf, true, "package should be marked isself")
  assert.Equal(t, pack2.IsLink, true, "package should be marked islink")
  assert.Equal(t, pack3.IsLink, true, "package should be marked islink")
  assert.NotEqual(t, pack2.LinkTarget, "", "package should have link target set")
  assert.NotEqual(t, pack3.LinkTarget, "", "package should have link target set")
*/
