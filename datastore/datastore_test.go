package datastore_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	gcpdatastore "cloud.google.com/go/datastore"

	"github.com/philippgille/gokv/datastore"
	"github.com/philippgille/gokv/encoding"
	"github.com/philippgille/gokv/test"
)

// TestClient tests if reading from, writing to and deleting from the store works properly.
// A struct is used as value. See TestTypes() for a test that is simpler but tests all types.
//
// Note: This test is only executed if the initial connection to Cloud Datastore works.
func TestClient(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to Cloud Datastore could be established. Probably not running in a proper test environment.")
	}

	// Test with JSON
	t.Run("JSON", func(t *testing.T) {
		client := createClient(t, encoding.JSON)
		defer client.Close()
		test.TestStore(client, t)
	})

	// Test with gob
	t.Run("gob", func(t *testing.T) {
		client := createClient(t, encoding.Gob)
		defer client.Close()
		test.TestStore(client, t)
	})
}

// TestTypes tests if setting and getting values works with all Go types.
//
// Note: This test is only executed if the initial connection to Cloud Datastore works.
func TestTypes(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to Cloud Datastore could be established. Probably not running in a proper test environment.")
	}

	// Test with JSON
	t.Run("JSON", func(t *testing.T) {
		client := createClient(t, encoding.JSON)
		defer client.Close()
		test.TestTypes(client, t)
	})

	// Test with gob
	t.Run("gob", func(t *testing.T) {
		client := createClient(t, encoding.Gob)
		defer client.Close()
		test.TestTypes(client, t)
	})
}

// TestClientConcurrent launches a bunch of goroutines that concurrently work with the Cloud Datastore client.
//
// Note: This test is only executed if the initial connection to Cloud Datastore works.
func TestClientConcurrent(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to Cloud Datastore could be established. Probably not running in a proper test environment.")
	}

	client := createClient(t, encoding.JSON)
	defer client.Close()

	// TODO: Should test 1000, but that only works with GCP
	// or a locally running emulator with enough resources.
	// It does NOT work on Travis CI (leads to timeouts within the goroutines).
	goroutineCount := 500

	test.TestConcurrentInteractions(t, goroutineCount, client)
}

// TestErrors tests some error cases.
//
// Note: This test is only executed if the initial connection to Cloud Datastore works.
func TestErrors(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to Cloud Datastore could be established. Probably not running in a proper test environment.")
	}

	// Test empty key
	client := createClient(t, encoding.JSON)
	defer client.Close()
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
//
// Note: This test is only executed if the initial connection to Cloud Datastore works.
func TestNil(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to Cloud Datastore could be established. Probably not running in a proper test environment.")
	}

	// Test setting nil

	t.Run("set nil with JSON marshalling", func(t *testing.T) {
		client := createClient(t, encoding.JSON)
		defer client.Close()
		err := client.Set("foo", nil)
		if err == nil {
			t.Error("Expected an error")
		}
	})

	t.Run("set nil with Gob marshalling", func(t *testing.T) {
		client := createClient(t, encoding.Gob)
		defer client.Close()
		err := client.Set("foo", nil)
		if err == nil {
			t.Error("Expected an error")
		}
	})

	// Test passing nil or pointer to nil value for retrieval

	createTest := func(codec encoding.Codec) func(t *testing.T) {
		return func(t *testing.T) {
			client := createClient(t, codec)
			defer client.Close()

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
//
// Note: This test is only executed if the initial connection to Cloud Datastore works.
func TestClose(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to Cloud Datastore could be established. Probably not running in a proper test environment.")
	}

	client := createClient(t, encoding.JSON)
	err := client.Close()
	if err != nil {
		t.Error(err)
	}
}

// checkConnection returns true if a connection could be made, false otherwise.
func checkConnection() bool {
	err := os.Setenv("DATASTORE_EMULATOR_HOST", "localhost:8081")
	if err != nil {
		fmt.Printf("Emulator environment variable couldn't be set: %v\n", err)
		return false
	}
	dsClient, err := gcpdatastore.NewClient(context.Background(), "gokv")
	if err != nil {
		fmt.Printf("Client couldn't be created: %v\n", err)
		return false
	}
	defer dsClient.Close()

	// Let's use AllocateIDs() as connection test.
	// It takes incomplete keys and returns valid keys.
	tctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	keys := []*gcpdatastore.Key{
		{
			Kind: "gokv",
		},
	}
	_, err = dsClient.AllocateIDs(tctx, keys)
	if err != nil {
		fmt.Printf("Connection attempt to Cloud Datastore failed: %v\n", err)
		return false
	}

	return true
}

func createClient(t *testing.T, codec encoding.Codec) datastore.Client {
	err := os.Setenv("DATASTORE_EMULATOR_HOST", "localhost:8081")
	if err != nil {
		t.Fatalf("Emulator environment variable couldn't be set: %v\n", err)
	}
	options := datastore.Options{
		ProjectID: "gokv",
		Codec:     codec,
	}
	client, err := datastore.NewClient(options)
	if err != nil {
		t.Fatal(err)
	}
	return client
}
