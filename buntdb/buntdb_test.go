package buntdb_test

import (
	"github.com/philippgille/gokv"
	"github.com/philippgille/gokv/buntdb"
	"github.com/philippgille/gokv/encoding"
	"github.com/philippgille/gokv/test"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

// TestStore tests if reading from, writing to and deleting from the store works properly.
// A struct is used as value. See TestTypes() for a test that is simpler but tests all types.
func TestStore(t *testing.T) {
	// Test with JSON
	t.Run("JSON", func(t *testing.T) {
		store, path := createStore(t, encoding.JSON)
		defer cleanUp(store, path)
		test.TestStore(store, t)
	})

	// Test with gob
	t.Run("gob", func(t *testing.T) {
		store, path := createStore(t, encoding.Gob)
		defer cleanUp(store, path)
		test.TestStore(store, t)
	})
}

// TestTypes tests if setting and getting values works with all Go types.
func TestTypes(t *testing.T) {
	// Test with JSON
	t.Run("JSON", func(t *testing.T) {
		store, path := createStore(t, encoding.JSON)
		defer cleanUp(store, path)
		test.TestTypes(store, t)
	})

	// Test with gob
	t.Run("gob", func(t *testing.T) {
		store, path := createStore(t, encoding.Gob)
		defer cleanUp(store, path)
		test.TestTypes(store, t)
	})
}

// TestStoreConcurrent launches a bunch of goroutines that concurrently work with one store.
// The store works with a single file, so everything should be locked properly.
func TestStoreConcurrent(t *testing.T) {
	store, path := createStore(t, encoding.JSON)
	defer cleanUp(store, path)

	goroutineCount := 1000

	test.TestConcurrentInteractions(t, goroutineCount, store)
}

// TestErrors tests some error cases.
func TestErrors(t *testing.T) {
	// Test empty key
	store, path := createStore(t, encoding.JSON)
	defer cleanUp(store, path)
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
		store, path := createStore(t, encoding.JSON)
		defer cleanUp(store, path)
		err := store.Set("foo", nil)
		if err == nil {
			t.Error("Expected an error")
		}
	})

	t.Run("set nil with Gob marshalling", func(t *testing.T) {
		store, path := createStore(t, encoding.Gob)
		defer cleanUp(store, path)
		err := store.Set("foo", nil)
		if err == nil {
			t.Error("Expected an error")
		}
	})

	// Test passing nil or pointer to nil value for retrieval

	createTest := func(codec encoding.Codec) func(t *testing.T) {
		return func(t *testing.T) {
			store, path := createStore(t, codec)
			defer cleanUp(store, path)

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
	store, path := createStore(t, encoding.JSON)
	defer os.RemoveAll(path)
	err := store.Close()
	if err != nil {
		t.Error(err)
	}
}

// TestNonExistingDir tests whether the implementation can create the given directory on its own.
func TestNonExistingDir(t *testing.T) {
	tmpDir := os.TempDir() + "/buntdb"
	err := os.RemoveAll(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	options := buntdb.Options{
		Path: tmpDir,
	}
	store, err := buntdb.NewStore(options)
	defer cleanUp(store, tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	err = store.Set("foo", "bar")
	if err != nil {
		t.Error(err)
	}
}

func createStore(t *testing.T, codec encoding.Codec) (buntdb.Store, string) {
	path := generateRandomTempDbPath(t)
	options := buntdb.Options{
		Path:  path,
		Codec: codec,
	}
	store, err := buntdb.NewStore(options)
	if err != nil {
		t.Fatal(err)
	}
	return store, path
}

func generateRandomTempDbPath(t *testing.T) string {
	path, err := ioutil.TempDir(os.TempDir(), "buntdb")
	if err != nil {
		t.Fatalf("Generating random DB path failed: %v", err)
	}
	path += "/buntdb.db"
	return path
}

// cleanUp cleans up (deletes) the database file that has been created during a test.
// If an error occurs the test is NOT marked as failed.
func cleanUp(store gokv.Store, path string) {
	err := store.Close()
	if err != nil {
		log.Printf("Error during cleaning up after a test (during closing the store): %v\n", err)
	}
	err = os.RemoveAll(path)
	if err != nil {
		log.Printf("Error during cleaning up after a test (during removing the data directory): %v\n", err)
	}
}
