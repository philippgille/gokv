package syncmap_test

import (
	"strconv"
	"sync"
	"testing"

	"github.com/philippgille/gokv/syncmap"
	"github.com/philippgille/gokv/test"
)

// TestGoMap tests if reading and writing to the store works properly.
func TestGoMap(t *testing.T) {
	goMap := syncmap.NewGoMap()

	test.TestStore(goMap, t)
}

// TestGoMapConcurrent launches a bunch of goroutines that concurrently work with one GoMap.
// The GoMap is a sync.Map, so the concurrency should be supported by the used package.
func TestGoMapConcurrent(t *testing.T) {
	goMap := syncmap.NewGoMap()

	goroutineCount := 1000

	waitGroup := sync.WaitGroup{}
	waitGroup.Add(goroutineCount) // Must be called before any goroutine is started
	for i := 0; i < goroutineCount; i++ {
		go test.InteractWithStore(goMap, strconv.Itoa(i), t, &waitGroup)
	}
	waitGroup.Wait()

	// Now make sure that all values are in the store
	expected := test.Foo{}
	for i := 0; i < goroutineCount; i++ {
		actualPtr := new(test.Foo)
		found, err := goMap.Get(strconv.Itoa(i), actualPtr)
		if err != nil {
			t.Errorf("An error occurred during the test: %v", err)
		}
		if !found {
			t.Errorf("No value was found, but should have been")
		}
		actual := *actualPtr
		if actual != expected {
			t.Errorf("Expected: %v, but was: %v", expected, actual)
		}
	}
}
