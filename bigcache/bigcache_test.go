package bigcache_test

import (
	"math/rand"
	"strconv"
	"testing"

	"github.com/philippgille/gokv/bigcache"
	"github.com/philippgille/gokv/encoding"
	"github.com/philippgille/gokv/test"
)

// TestStore tests if reading and writing to the store works properly.
func TestStore(t *testing.T) {
	// Test with JSON
	t.Run("JSON", func(t *testing.T) {
		store := createStore(t, encoding.JSON)
		defer store.Close()
		test.TestStore(store, t)
	})

	// Test with gob
	t.Run("gob", func(t *testing.T) {
		store := createStore(t, encoding.Gob)
		defer store.Close()
		test.TestStore(store, t)
	})
}

// TestTypes tests if setting and getting values works with all Go types.
func TestTypes(t *testing.T) {
	// Test with JSON
	t.Run("JSON", func(t *testing.T) {
		store := createStore(t, encoding.JSON)
		defer store.Close()
		test.TestTypes(store, t)
	})

	// Test with gob
	t.Run("gob", func(t *testing.T) {
		store := createStore(t, encoding.Gob)
		defer store.Close()
		test.TestTypes(store, t)
	})
}

// TestStoreConcurrent launches a bunch of goroutines that concurrently work with one store.
func TestStoreConcurrent(t *testing.T) {
	store := createStore(t, encoding.JSON)
	defer store.Close()

	goroutineCount := 1000

	test.TestConcurrentInteractions(t, goroutineCount, store)
}

// TestErrors tests some error cases.
func TestErrors(t *testing.T) {
	// Test empty key
	store := createStore(t, encoding.JSON)
	defer store.Close()
	err := store.Set("", "bar")
	if err == nil {
		t.Error("Expected an error")
	}
	_, err = store.Get("", new(string))
	if err == nil {
		t.Error("Expected an error")
	}
	err = store.Delete("")
	if err == nil {
		t.Error("Expected an error")
	}
}

// TestNil tests the behaviour when passing nil or pointers to nil values to some methods.
func TestNil(t *testing.T) {
	// Test setting nil

	t.Run("set nil with JSON marshalling", func(t *testing.T) {
		store := createStore(t, encoding.JSON)
		defer store.Close()
		err := store.Set("foo", nil)
		if err == nil {
			t.Error("Expected an error")
		}
	})

	t.Run("set nil with Gob marshalling", func(t *testing.T) {
		store := createStore(t, encoding.Gob)
		defer store.Close()
		err := store.Set("foo", nil)
		if err == nil {
			t.Error("Expected an error")
		}
	})

	// Test passing nil or pointer to nil value for retrieval

	createTest := func(codec encoding.Codec) func(t *testing.T) {
		return func(t *testing.T) {
			store := createStore(t, codec)
			defer store.Close()

			// Prep
			err := store.Set("foo", test.Foo{Bar: "baz"})
			if err != nil {
				t.Error(err)
			}

			_, err = store.Get("foo", nil) // actually nil
			if err == nil {
				t.Error("An error was expected")
			}

			var i interface{} // actually nil
			_, err = store.Get("foo", i)
			if err == nil {
				t.Error("An error was expected")
			}

			var valPtr *test.Foo // nil value
			_, err = store.Get("foo", valPtr)
			if err == nil {
				t.Error("An error was expected")
			}
		}
	}
	t.Run("get with nil / nil value parameter", createTest(encoding.JSON))
	t.Run("get with nil / nil value parameter", createTest(encoding.Gob))
}

// TestClose tests if the close method returns any errors.
func TestClose(t *testing.T) {
	store := createStore(t, encoding.JSON)
	err := store.Close()
	if err != nil {
		t.Error(err)
	}
}

// TestEvictionOnMaxSize tests if entries are evicted when the max size is reached when NO eviction time is set.
func TestEvictionOnMaxSize(t *testing.T) {
	// Test with small max size (1 MiB) and eviction of 0
	options := bigcache.Options{
		HardMaxCacheSize: 1,
	}
	store, err := bigcache.NewStore(options)
	defer store.Close()
	if err != nil {
		t.Fatal(err)
	}

	// Save 1*1024*1024 entries that are at least 1 byte each.
	// This should lead to reaching the 1 MiB limit.
	// Note: Storing "foo" 1024*1024 times isn't enough sometimes.
	count := 1 * 1024 * 1024
	for i := 0; i < count && err == nil; i++ {
		err = store.Set(strconv.Itoa(i), strconv.Itoa(rand.Int()))
		if err != nil {
			t.Error(err)
		}
	}

	// The first entry shouldn't exist anymore, because it was evicted due to the max size, NOT because of some eviction time.
	valPtr := new(string)
	found, err := store.Get("1", valPtr)
	if err != nil {
		t.Error(err)
	}
	if found {
		t.Error("First value should have been evicted, but wasn't")
	}
}

func createStore(t *testing.T, codec encoding.Codec) bigcache.Store {
	options := bigcache.Options{
		Codec: codec,
	}
	store, err := bigcache.NewStore(options)
	if err != nil {
		t.Fatal(err)
	}
	return store
}
