package main

import (
	"fmt"
	"os"
	"path"

	"github.com/juju/errors"
)

func setVendorEnv() error {
	dir, err := os.Getwd()

	if err != nil {
		return errors.Trace(err)
	}

	newGoPath := path.Join(dir, ".vendor")
	newPath := fmt.Sprintf("%s:%s", path.Join(dir, ".vendor", "bin"), InitialPath)

	err = os.Setenv("PATH", newPath)
	if err != nil {
		return errors.Trace(err)
	}

	err = os.Setenv("GOPATH", newGoPath)
	if err != nil {
		return errors.Trace(err)
	}

	return nil
}

func unsetVendorEnv() error {
	err := os.Setenv("PATH", InitialPath)
	if err != nil {
		return errors.Trace(err)
	}

	err = os.Setenv("GOPATH", InitialGoPath)
	if err != nil {
		return errors.Trace(err)
	}

	return nil
}
