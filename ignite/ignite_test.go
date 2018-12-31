package ignite_test

import (
	"log"
	"net"
	"testing"
	"time"

	orig "github.com/amsokol/ignite-go-client/binary/v1"

	"github.com/philippgille/gokv/encoding"
	"github.com/philippgille/gokv/ignite"
	"github.com/philippgille/gokv/test"
)

// TestClient tests if reading from, writing to and deleting from the store works properly.
// A struct is used as value. See TestTypes() for a test that is simpler but tests all types.
//
// Note: This test is only executed if the initial connection to Apache Ignite works.
func TestClient(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to Apache Ignite could be established. Probably not running in a proper test environment.")
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
// Note: This test is only executed if the initial connection to Apache Ignite works.
func TestTypes(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to Apache Ignite could be established. Probably not running in a proper test environment.")
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

// TestClientConcurrent launches a bunch of goroutines that concurrently work with the Apache Ignite client.
//
// Note: This test is only executed if the initial connection to Apache Ignite works.
func TestClientConcurrent(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to Apache Ignite could be established. Probably not running in a proper test environment.")
	}

	client := createClient(t, encoding.JSON)
	defer client.Close()

	goroutineCount := 1000

	test.TestConcurrentInteractions(t, goroutineCount, client)
}

// TestErrors tests some error cases.
//
// Note: This test is only executed if the initial connection to Apache Ignite works.
func TestErrors(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to Apache Ignite could be established. Probably not running in a proper test environment.")
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
// Note: This test is only executed if the initial connection to Apache Ignite works.
func TestNil(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to Apache Ignite could be established. Probably not running in a proper test environment.")
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
	t.Run("get with nil / nil value parameter", createTest(encoding.JSON))
	t.Run("get with nil / nil value parameter", createTest(encoding.Gob))
}

// TestClose tests if the close method returns any errors.
//
// Note: This test is only executed if the initial connection to Apache Ignite works.
func TestClose(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to Apache Ignite could be established. Probably not running in a proper test environment.")
	}

	client := createClient(t, encoding.JSON)
	err := client.Close()
	if err != nil {
		t.Error(err)
	}
}

// checkConnection returns true if a connection could be made, false otherwise.
func checkConnection() bool {
	connInfo := orig.ConnInfo{
		Dialer: net.Dialer{
			Timeout: 2 * time.Second,
		},
		Host:    "localhost",
		Major:   1,
		Minor:   1,
		Network: "tcp",
		Port:    10800,
		// Go zero values for Username, Password, Patch and TLSConfig.
	}
	c, err := orig.Connect(connInfo)
	if err != nil {
		log.Printf("Couldn't connect during checkConnection(). Error: %v\n", err)
		return false
	}
	defer c.Close()
	return c.Connected()
}

func createClient(t *testing.T, codec encoding.Codec) ignite.Client {
	options := ignite.Options{}
	client, err := ignite.NewClient(options)
	if err != nil {
		t.Fatal(err)
	}
	return client
}
