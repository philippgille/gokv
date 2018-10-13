package gokv_test

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"

	"github.com/philippgille/ln-paywall/wall"
)

type foo struct {
	Bar string
}

// testStorageClient tests if reading from and writing to the storage works properly.
func testStorageClient(storageClient wall.StorageClient, t *testing.T) {
	key := strconv.FormatInt(rand.Int63(), 10)

	// Initially the key shouldn't exist
	found, err := storageClient.Get(key, new(foo))
	if err != nil {
		t.Error(err)
	}
	if found {
		t.Errorf("A value was found, but no value was expected")
	}

	// Store an object
	val := foo{
		Bar: "baz",
	}
	err = storageClient.Set(key, val)
	if err != nil {
		t.Error(err)
	}

	// Retrieve the object
	expected := val
	actualPtr := new(foo)
	found, err = storageClient.Get(key, actualPtr)
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
}

// interactWithStorage reads from and writes to the DB. Meant to be executed in a goroutine.
// Does NOT check if the DB works correctly (that's done elsewhere),
// only checks for errors that might occur due to concurrent access.
func interactWithStorage(storageClient wall.StorageClient, key string, t *testing.T, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()

	// Read
	_, err := storageClient.Get(key, new(foo))
	if err != nil {
		t.Error(err)
	}
	// Write
	err = storageClient.Set(key, foo{})
	if err != nil {
		t.Error(err)
	}
	// Read
	_, err = storageClient.Get(key, new(foo))
	if err != nil {
		t.Error(err)
	}
}
