package sqlite_test

import (
	"log"
	"os"
	"testing"

	"github.com/philippgille/gokv/encoding"
	"github.com/philippgille/gokv/sqlite"
	"github.com/philippgille/gokv/test"
)

// TestClient tests if reading from, writing to and deleting from the store works properly.
// A struct is used as value. See TestTypes() for a test that is simpler but tests all types.
func TestClient(t *testing.T) {
	// Test with JSON
	t.Run("JSON", func(t *testing.T) {
		client := createClient(t, encoding.JSON)
		defer cleanUp(client, "potato.db")
		test.TestStore(client, t)
	})

	// Test with gob
	t.Run("gob", func(t *testing.T) {
		client := createClient(t, encoding.Gob)
		defer cleanUp(client, "potato.db")
		test.TestStore(client, t)
	})
}

// TestTypes tests if setting and getting values works with all Go types.
func TestTypes(t *testing.T) {
	// Test with JSON
	t.Run("JSON", func(t *testing.T) {
		client := createClient(t, encoding.JSON)
		defer cleanUp(client, "potato.db")
		test.TestTypes(client, t)
	})

	// Test with gob
	t.Run("gob", func(t *testing.T) {
		client := createClient(t, encoding.Gob)
		defer cleanUp(client, "potato.db")
		test.TestTypes(client, t)
	})
}

// TestClientConcurrent launches a bunch of goroutines that concurrently work with the sqlite client.
func TestClientConcurrent(t *testing.T) {
	client := createClient(t, encoding.JSON)
	defer cleanUp(client, "potato.db")
	goroutineCount := 1000

	test.TestConcurrentInteractions(t, goroutineCount, client)
}

// TestErrors tests some error cases.
func TestErrors(t *testing.T) {
	// Test empty key
	client := createClient(t, encoding.JSON)
	defer cleanUp(client, "potato.db")
	err := client.Set("", "bar")
	if err == nil {
		t.Error("Expected an error")
	}
	_, err = client.Get("", new(string))
	if err == nil {
		t.Error("Expected an error")
	}
	err = client.Delete("")
	if err == nil {
		t.Error("Expected an error")
	}
}

// TestNil tests the behaviour when passing nil or pointers to nil values to some methods.
func TestNil(t *testing.T) {
	// Test setting nil

	t.Run("set nil with JSON marshalling", func(t *testing.T) {
		client := createClient(t, encoding.JSON)
		defer cleanUp(client, "potato.db")
		err := client.Set("foo", nil)
		if err == nil {
			t.Error("Expected an error")
		}
	})

	t.Run("set nil with Gob marshalling", func(t *testing.T) {
		client := createClient(t, encoding.Gob)
		defer cleanUp(client, "potato.db")
		err := client.Set("foo", nil)
		if err == nil {
			t.Error("Expected an error")
		}
	})

	// Test passing nil or pointer to nil value for retrieval

	createTest := func(codec encoding.Codec) func(t *testing.T) {
		return func(t *testing.T) {
			client := createClient(t, codec)
			defer cleanUp(client, "potato.db")

			// Prep
			err := client.Set("foo", test.Foo{Bar: "baz"})
			if err != nil {
				t.Error(err)
			}

			_, err = client.Get("foo", nil) // actually nil
			if err == nil {
				t.Error("An error was expected")
			}

			var i any // actually nil
			_, err = client.Get("foo", i)
			if err == nil {
				t.Error("An error was expected")
			}

			var valPtr *test.Foo // nil value
			_, err = client.Get("foo", valPtr)
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
	client := createClient(t, encoding.JSON)
	defer cleanUp(client, "potato.db")
	err := client.Close()
	if err != nil {
		t.Error(err)
	}
}

// checkConnection returns true if a connection could be made, false otherwise.
func TestCheckConnection(t *testing.T) {
	client, err := sqlite.NewClient(sqlite.Options{
		Path:  "potato.db",
		Codec: encoding.JSON,
	})
	if err != nil {
		log.Printf("An error occurred during testing the connection to the server: %v\n", err)
		t.Error(err)
	}
	defer cleanUp(client, "potato.db")

	err = client.C.Ping()
	if err != nil {
		log.Printf("An error occurred during testing the connection to the server: %v\n", err)
		t.Error(err)
	}
}

func createClient(t *testing.T, codec encoding.Codec) sqlite.Client {
	options := sqlite.Options{
		Path:  "potato.db",
		Codec: codec,
	}
	client, err := sqlite.NewClient(options)
	if err != nil {
		t.Fatal(err)
	}
	return client
}

// cleanUp cleans up (deletes) the database file that has been created during a test.
// If an error occurs the test is NOT marked as failed.
func cleanUp(client sqlite.Client, path string) {
	err := client.Close()
	if err != nil {
		log.Printf("Error during cleaning up after a test (during closing the store): %v\n", err)
	}
	err = os.RemoveAll(path)
	if err != nil {
		log.Printf("Error during cleaning up after a test (during removing the data directory): %v\n", err)
	}
}
