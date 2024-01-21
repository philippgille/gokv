package consul_test

import (
	"log"
	"strconv"
	"testing"
	"time"

	"github.com/hashicorp/consul/api"

	"github.com/philippgille/gokv/consul"
	"github.com/philippgille/gokv/encoding"
	"github.com/philippgille/gokv/test"
)

// TestClient tests if reading from, writing to and deleting from the store works properly.
// A struct is used as value. See TestTypes() for a test that is simpler but tests all types.
func TestClient(t *testing.T) {
	// Test with JSON
	t.Run("JSON", func(t *testing.T) {
		client := createClient(t, encoding.JSON)
		test.TestStore(client, t)
	})

	// Test with gob
	t.Run("gob", func(t *testing.T) {
		client := createClient(t, encoding.Gob)
		test.TestStore(client, t)
	})
}

// TestTypes tests if setting and getting values works with all Go types.
func TestTypes(t *testing.T) {
	// Test with JSON
	t.Run("JSON", func(t *testing.T) {
		client := createClient(t, encoding.JSON)
		test.TestTypes(client, t)
	})

	// Test with gob
	t.Run("gob", func(t *testing.T) {
		client := createClient(t, encoding.Gob)
		test.TestTypes(client, t)
	})
}

// TestClientConcurrent launches a bunch of goroutines that concurrently work with the Consul client.
func TestClientConcurrent(t *testing.T) {
	client := createClient(t, encoding.JSON)

	// With 1000 we get rate limiting errors from the Consul container,
	// despite setting CONSUL_LOCAL_CONFIG='{"limits":{"http_max_conns_per_client":1000}}'.
	goroutineCount := 500

	test.TestConcurrentInteractions(t, goroutineCount, client)
}

// TestErrors tests some error cases.
func TestErrors(t *testing.T) {
	// Test empty key
	client := createClient(t, encoding.JSON)
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
		err := client.Set("foo", nil)
		if err == nil {
			t.Error("Expected an error")
		}
	})

	t.Run("set nil with Gob marshalling", func(t *testing.T) {
		client := createClient(t, encoding.Gob)
		err := client.Set("foo", nil)
		if err == nil {
			t.Error("Expected an error")
		}
	})

	// Test passing nil or pointer to nil value for retrieval

	createTest := func(codec encoding.Codec) func(t *testing.T) {
		return func(t *testing.T) {
			client := createClient(t, codec)

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
	err := client.Close()
	if err != nil {
		t.Error(err)
	}
}

// checkConnection returns true if a connection could be made, false otherwise.
func checkConnection() bool {
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		log.Printf("An error occurred during testing the connection to the server: %v\n", err)
		return false
	}
	res, err := client.Status().Leader()
	if err != nil {
		log.Printf("An error occurred during testing the connection to the server: %v\n", err)
		return false
	} else if res == "" {
		log.Println("Status().Leader() response is empty")
		return false
	}
	return true
}

func createClient(t *testing.T, codec encoding.Codec) consul.Client {
	options := consul.Options{}
	options.Folder = "test_" + strconv.FormatInt(time.Now().Unix(), 10)
	options.Codec = codec
	client, err := consul.NewClient(options)
	if err != nil {
		t.Fatal(err)
	}
	return client
}
