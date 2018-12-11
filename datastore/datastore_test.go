package datastore_test

import (
	"context"
	"fmt"
	"testing"

	gcpdatastore "cloud.google.com/go/datastore"

	"github.com/philippgille/gokv/datastore"
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
		client := createClient(t, datastore.JSON)
		test.TestStore(client, t)
	})

	// Test with gob
	t.Run("gob", func(t *testing.T) {
		client := createClient(t, datastore.Gob)
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
		client := createClient(t, datastore.JSON)
		test.TestTypes(client, t)
	})

	// Test with gob
	t.Run("gob", func(t *testing.T) {
		client := createClient(t, datastore.Gob)
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

	client := createClient(t, datastore.JSON)

	goroutineCount := 1000

	test.TestConcurrentInteractions(t, goroutineCount, client)
}

// TestErrors tests some error cases.
//
// Note: This test is only executed if the initial connection to Cloud Datastore works.
func TestErrors(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to Cloud Datastore could be established. Probably not running in a proper test environment.")
	}

	// Test with a bad MarshalFormat enum value

	client := createClient(t, datastore.MarshalFormat(19))
	err := client.Set("foo", "bar")
	if err == nil {
		t.Error("An error should have occurred, but didn't")
	}
	// TODO: store some value for "foo", so retrieving the value works.
	// Just the unmarshalling should fail.
	// _, err = client.Get("foo", new(string))
	// if err == nil {
	// 	t.Error("An error should have occurred, but didn't")
	// }

	// Test empty key
	err = client.Set("", "bar")
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
		client := createClient(t, datastore.JSON)
		err := client.Set("foo", nil)
		if err == nil {
			t.Error("Expected an error")
		}
	})

	t.Run("set nil with Gob marshalling", func(t *testing.T) {
		client := createClient(t, datastore.Gob)
		err := client.Set("foo", nil)
		if err == nil {
			t.Error("Expected an error")
		}
	})

	// Test passing nil or pointer to nil value for retrieval

	createTest := func(mf datastore.MarshalFormat) func(t *testing.T) {
		return func(t *testing.T) {
			client := createClient(t, mf)

			// Prep
			err := client.Set("foo", test.Foo{Bar: "baz"})
			if err != nil {
				t.Error(err)
			}

			_, err = client.Get("foo", nil) // actually nil
			if err == nil {
				t.Error("An error was expected")
			}

			var i interface{} // actually nil
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
	t.Run("get with nil / nil value parameter", createTest(datastore.JSON))
	t.Run("get with nil / nil value parameter", createTest(datastore.Gob))
}

// TestClose tests if the close method returns any errors.
//
// Note: This test is only executed if the initial connection to Cloud Datastore works.
func TestClose(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to Cloud Datastore could be established. Probably not running in a proper test environment.")
	}

	client := createClient(t, datastore.JSON)
	err := client.Close()
	if err != nil {
		t.Error(err)
	}
}

// checkConnection returns true if a connection could be made, false otherwise.
func checkConnection() bool {
	// TODO: os.Setenv("DATASTORE_EMULATOR_HOST", "?")
	dsClient, err := gcpdatastore.NewClient(context.Background(), "gokv")
	if err != nil {
		fmt.Printf("Client couldn't be created: %v\n", err)
		return false
	}

	// Let's use AllocateIDs() as connection test.
	// It takes incomplete keys and returns valid keys.
	keys := []*gcpdatastore.Key{
		&gcpdatastore.Key{
			Kind: "gokv",
		},
	}
	_, err = dsClient.AllocateIDs(context.Background(), keys)
	if err != nil {
		fmt.Printf("Connection attempt to Cloud Datastore failed: %v\n", err)
		return false
	}

	return true
}

func createClient(t *testing.T, mf datastore.MarshalFormat) datastore.Client {
	// TODO: os.Setenv("DATASTORE_EMULATOR_HOST", "?")
	options := datastore.Options{
		ProjectID:     "gokv",
		MarshalFormat: mf,
	}
	client, err := datastore.NewClient(options)
	if err != nil {
		t.Fatal(err)
	}
	return client
}
