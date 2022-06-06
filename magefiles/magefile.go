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
		return sh.Run("./build/update-deps.ps1")
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
		return sh.Run("./build/build.ps1")
	case "darwin":
		fallthrough
	case "linux":
		return sh.Run("./build/build.sh")
	}
	return errors.New("your OS is not supported")
}

// Test tests all modules.
func Test() error {
	switch runtime.GOOS {
	// TODO: Support Windows. Instead of writing a test.ps1, implement it here in the magefile to also replace the test.sh.
	case "darwin":
		fallthrough
	case "linux":
		return sh.Run("./build/test.sh")
	}
	return errors.New("your OS is not supported")
}

// Clean cleans the build/test output, like coverage.txt files
func Clean() error {
	p := script.FindFiles(".").
		Match("coverage.txt").
		ExecForEach("rm ./{{.}}") // On Windows `rm` works as it's an alias for Remove-Item
	p.Wait()
	return p.Error()
}
