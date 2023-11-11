//go:build mage

package main

import (
	"errors"
	"runtime"

	"github.com/bitfield/script"
	"github.com/magefile/mage/sh"
)

// Update updates the dependencies of all modules.
// It only updates direct dependencies within the same major version, like `go get` with `@latest` does.
// It doesn't update transitive dependencies, like `go get -u` does.
// It also runs `go mod tidy` for all modules after updating.
func Update() error {
	switch runtime.GOOS {
	case "windows":
		return sh.Run("pwsh.exe", "./build/update-deps.ps1")
	case "darwin":
		fallthrough
	case "linux":
		return sh.Run("./build/update-deps.sh")
	}
	return errors.New("your OS is not supported")
}

// Build builds all modules.
func Build() error {
	switch runtime.GOOS {
	case "windows":
		return sh.Run("pwsh.exe", "./build/build.ps1")
	case "darwin":
		fallthrough
	case "linux":
		return sh.Run("./build/build.sh")
	}
	return errors.New("your OS is not supported")
}

// Test tests the given module. Pass "all" to test all modules.
func Test(module string) error {
	if module == "all" {
		// Helper packages and examples currently don't have tests, so currently for *all* tests we can just iterate all `gokv.Store` implementations
		// TODO: Add tests for helper and example packages, then change this behavior.
		impls, err := script.File("./build/implementations").Slice()
		if err != nil {
			return err
		}
		for _, impl := range impls {
			err = testImpl(impl)
			if err != nil {
				return err
			}
		}
		return nil
	}

	switch module {
	case "encoding", "sql", "test", "util":
		return errors.New("module " + module + " doesn't have any tests")
	case "examples":
		return errors.New("examples don't have any tests")
	}

	i, err := script.File("./build/implementations").Match(module).CountLines()
	if err != nil {
		return err
	}
	if i == 0 {
		return errors.New("module from parameter not found")
	}

	return testImpl(module)
}

// Clean cleans the build/test output, like coverage.txt files
func Clean() error {
	p := script.FindFiles(".").
		Match("coverage.txt").
		ExecForEach("rm ./{{.}}") // On Windows `rm` works as it's an alias for Remove-Item
	p.Wait()
	return p.Error()
}
