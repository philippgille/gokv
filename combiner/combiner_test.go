package combiner_test

import (
	"testing"

	"github.com/philippgille/gokv/combiner"
	"github.com/philippgille/gokv/file"
	"github.com/philippgille/gokv/gomap"
)

// TestCombiner tests the basic functionality of the combiner.
func TestCombiner(t *testing.T) {
	gomapStore := gomap.NewStore(gomap.DefaultOptions)
	fileStore, err := file.NewStore(file.DefaultOptions)
	if err != nil {
		t.Fatal(err)
	}
	combiner, err := combiner.NewCombiner(combiner.DefaultOptions, gomapStore, fileStore)
	if err != nil {
		t.Fatal(err)
	}
	defer combiner.Close()

	err = combiner.Set("foo", "bar")
	if err != nil {
		t.Error(err)
	}

	var result string
	found, err := combiner.Get("foo", &result)
	if err != nil {
		t.Error(err)
	}
	if !found {
		t.Error("Value not found")
	}

	if result != "bar" {
		t.Errorf(`Expected "bar", but got %v`, result)
	}
}
