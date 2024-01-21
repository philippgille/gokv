package memcached_test

import (
	"log"
	"testing"
	"time"

	"github.com/bradfitz/gomemcache/memcache"

	"github.com/philippgille/gokv/encoding"
	"github.com/philippgille/gokv/memcached"
	"github.com/philippgille/gokv/test"
)

// TestClient tests if reading from, writing to and deleting from the store works properly.
// A struct is used as value. See TestTypes() for a test that is simpler but tests all types.
//
// Note: This test is only executed if the initial connection to Memcached works.
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
//
// Note: This test is only executed if the initial connection to Memcached works.
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

// TestClientConcurrent launches a bunch of goroutines that concurrently work with the Memcached client.
//
// Note: This test is only executed if the initial connection to Memcached works.
func TestClientConcurrent(t *testing.T) {
	client := createClient(t, encoding.JSON)

	// TODO: 1000 leads to timeout errors every time.
	// Looks like the server load is too high, but should that really be the case with Memcached?
	goroutineCount := 250

	test.TestConcurrentInteractions(t, goroutineCount, client)
}

// TestErrors tests some error cases.
//
// Note: This test is only executed if the initial connection to Memcached works.
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
//
// Note: This test is only executed if the initial connection to Memcached works.
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
//
// Note: This test is only executed if the initial connection to Memcached works.
func TestClose(t *testing.T) {
	client := createClient(t, encoding.JSON)
	err := client.Close()
	if err != nil {
		t.Error(err)
	}
}

// TestDefaultTimeout tests if the client works with the default timeout.
// Currently, the createClient() method is used in other tests,
// which sets the timeout to 2 seconds due to errors during the concurrency test.
//
// Note: This test is only executed if the initial connection to Memcached works.
func TestDefaultTimeout(t *testing.T) {
	options := memcached.Options{}
	client, err := memcached.NewClient(options)
	if err != nil {
		t.Error(err)
	}

	err = client.Set("foo", "bar")
	if err != nil {
		t.Error(err)
	}
	vPtr := new(string)
	found, err := client.Get("foo", vPtr)
	if err != nil {
		t.Error(err)
	}
	if !found {
		t.Error("A value should have been found, but wasn't.")
	}
	if *vPtr != "bar" {
		t.Errorf("Expectec %v, but was %v", "bar", *vPtr)
	}
}

// checkConnection returns true if a connection could be made, false otherwise.
func checkConnection() bool {
	mc := memcache.New("localhost:11211")
	_, err := mc.Get("foo")
	if err == nil || err == memcache.ErrCacheMiss {
		return true
	}
	log.Printf("An error occurred during testing the connection to the server: %v\n", err)
	return false
}

func createClient(t *testing.T, codec encoding.Codec) memcached.Client {
	// TODO: High timeout is necessary for local testing to avoid timeout errors,
	// but 2 seconds seem way too high.
	timeout := 2 * time.Second
	options := memcached.Options{
		Timeout: &timeout,
		Codec:   codec,
	}
	client, err := memcached.NewClient(options)
	if err != nil {
		t.Fatal(err)
	}
	return client
}
