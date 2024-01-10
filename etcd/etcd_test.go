package etcd_test

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/philippgille/gokv/encoding"
	"github.com/philippgille/gokv/etcd"
	"github.com/philippgille/gokv/test"
)

// TestClient tests if reading from, writing to and deleting from the store works properly.
// A struct is used as value. See TestTypes() for a test that is simpler but tests all types.
//
// Note: This test is only executed if the initial connection to etcd works.
func TestClient(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to etcd could be established. Probably not running in a proper test environment.")
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
// Note: This test is only executed if the initial connection to etcd works.
func TestTypes(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to etcd could be established. Probably not running in a proper test environment.")
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

// TestClientConcurrent launches a bunch of goroutines that concurrently work with the etcd client.
//
// Note: This test is only executed if the initial connection to etcd works.
func TestClientConcurrent(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to etcd could be established. Probably not running in a proper test environment.")
	}

	client := createClient(t, encoding.JSON)
	defer client.Close()

	// The etcd server sometimes has issues with this test, but only in the CI environment.
	// Locally it works fine.
	var goroutineCount int
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		goroutineCount = 200
	} else {
		goroutineCount = 1000
	}

	test.TestConcurrentInteractions(t, goroutineCount, client)
}

// TestErrors tests some error cases.
//
// Note: This test is only executed if the initial connection to etcd works.
func TestErrors(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to etcd could be established. Probably not running in a proper test environment.")
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
// Note: This test is only executed if the initial connection to etcd works.
func TestNil(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to etcd could be established. Probably not running in a proper test environment.")
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
// Note: This test is only executed if the initial connection to etcd works.
func TestClose(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to etcd could be established. Probably not running in a proper test environment.")
	}

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
	if !checkConnection() {
		t.Skip("No connection to etcd could be established. Probably not running in a proper test environment.")
	}

	options := etcd.Options{}
	client, err := etcd.NewClient(options)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

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
	// clientv3.New() should block when a DialTimeout is set,
	// according to https://github.com/etcd-io/etcd/issues/9829.
	// TODO: But it doesn't.
	// cli, err := clientv3.NewFromURL("localhost:2379")
	config := clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 2 * time.Second,
	}

	cli, err := clientv3.New(config)
	if err != nil {
		log.Printf("An error occurred during testing the connection to the server: %v\n", err)
		return false
	}
	defer cli.Close()

	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	statusRes, err := cli.Status(ctxWithTimeout, "localhost:2379")
	if err != nil {
		log.Printf("An error occurred during testing the connection to the server: %v\n", err)
		return false
	} else if statusRes == nil {
		return false
	}
	return true
}

func createClient(t *testing.T, codec encoding.Codec) etcd.Client {
	timeout := 2 * time.Second
	options := etcd.Options{
		Timeout: &timeout,
		Codec:   codec,
	}
	client, err := etcd.NewClient(options)
	if err != nil {
		t.Fatal(err)
	}
	return client
}
