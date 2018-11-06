package badgerdb_test

import (
	"io/ioutil"
	"os"
	"strconv"
	"sync"
	"testing"

	"github.com/philippgille/gokv/badgerdb"
	"github.com/philippgille/gokv/test"
)

// TestStore tests if reading and writing to the store works properly.
func TestStore(t *testing.T) {
	options := badgerdb.Options{
		Dir: generateRandomTempDBpath(t),
	}
	store, err := badgerdb.NewStore(options)
	if err != nil {
		t.Error(err)
	}

	test.TestStore(store, t)
}

// TestStoreConcurrent launches a bunch of goroutines that concurrently work with one store.
// The store works with a single file, so everything should be locked properly.
// The locking is implemented in the BadgerDB package, but test it nonetheless.
func TestStoreConcurrent(t *testing.T) {
	options := badgerdb.Options{
		Dir: generateRandomTempDBpath(t),
	}
	store, err := badgerdb.NewStore(options)
	if err != nil {
		t.Error(err)
	}

	goroutineCount := 1000

	waitGroup := sync.WaitGroup{}
	waitGroup.Add(goroutineCount) // Must be called before any goroutine is started
	for i := 0; i < goroutineCount; i++ {
		go test.InteractWithStore(store, strconv.Itoa(i), t, &waitGroup)
	}
	waitGroup.Wait()

	// Now make sure that all values are in the store
	expected := test.Foo{}
	for i := 0; i < goroutineCount; i++ {
		actualPtr := new(test.Foo)
		found, err := store.Get(strconv.Itoa(i), actualPtr)
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

func generateRandomTempDBpath(t *testing.T) string {
	path, err := ioutil.TempDir(os.TempDir(), "BadgerDB")
	if err != nil {
		t.Errorf("Generating random DB path failed: %v", err)
	}
	return path
}
