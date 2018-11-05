package test

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"

	"github.com/philippgille/gokv"
)

// Foo is just some struct for common tests.
type Foo struct {
	Bar string
}

// TestStore tests if reading from and writing to the store works properly.
func TestStore(store gokv.Store, t *testing.T) {
	key := strconv.FormatInt(rand.Int63(), 10)

	// Initially the key shouldn't exist
	found, err := store.Get(key, new(Foo))
	if err != nil {
		t.Error(err)
	}
	if found {
		t.Error("A value was found, but no value was expected")
	}

	// Store an object
	val := Foo{
		Bar: "baz",
	}
	err = store.Set(key, val)
	if err != nil {
		t.Error(err)
	}

	// Retrieve the object
	expected := val
	actualPtr := new(Foo)
	found, err = store.Get(key, actualPtr)
	if err != nil {
		t.Error(err)
	}
	if !found {
		t.Errorf("No value was found, but should have been")
	}
	actual := *actualPtr
	if actual != expected {
		t.Errorf("Expected: %v, but was: %v", expected, actual)
	}

	// Delete
	err = store.Delete(key)
	if err != nil {
		t.Error(err)
	}
	// Key-value pair shouldn't exist anymore
	found, err = store.Get(key, new(Foo))
	if err != nil {
		t.Error(err)
	}
	if found {
		t.Error("A value was found, but no value was expected")
	}
}

// InteractWithStore reads from and writes to the DB. Meant to be executed in a goroutine.
// Does NOT check if the DB works correctly (that's done elsewhere),
// only checks for errors that might occur due to concurrent access.
func InteractWithStore(store gokv.Store, key string, t *testing.T, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()

	// Read
	_, err := store.Get(key, new(Foo))
	if err != nil {
		t.Error(err)
	}
	// Write
	err = store.Set(key, Foo{})
	if err != nil {
		t.Error(err)
	}
	// Read
	_, err = store.Get(key, new(Foo))
	if err != nil {
		t.Error(err)
	}
}
