package etcd_test

import (
	"context"
	"log"
	"strconv"
	"sync"
	"testing"
	"time"

	"go.etcd.io/etcd/clientv3"

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
		client := createClient(t, etcd.JSON)
		test.TestStore(client, t)
	})

	// Test with gob
	t.Run("gob", func(t *testing.T) {
		client := createClient(t, etcd.Gob)
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
		client := createClient(t, etcd.JSON)
		test.TestTypes(client, t)
	})

	// Test with gob
	t.Run("gob", func(t *testing.T) {
		client := createClient(t, etcd.Gob)
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

	client := createClient(t, etcd.JSON)

	goroutineCount := 1000

	waitGroup := sync.WaitGroup{}
	waitGroup.Add(goroutineCount) // Must be called before any goroutine is started
	for i := 0; i < goroutineCount; i++ {
		go test.InteractWithStore(client, strconv.Itoa(i), t, &waitGroup)
	}
	waitGroup.Wait()

	// Now make sure that all values are in the store
	expected := test.Foo{}
	for i := 0; i < goroutineCount; i++ {
		actualPtr := new(test.Foo)
		found, err := client.Get(strconv.Itoa(i), actualPtr)
		if err != nil {
			t.Errorf("An error occurred during the test: %v", err)
		}
		if !found {
			t.Error("No value was found, but should have been")
		}
		actual := *actualPtr
		if actual != expected {
			t.Errorf("Expected: %v, but was: %v", expected, actual)
		}
	}
}

// TestErrors tests some error cases.
//
// Note: This test is only executed if the initial connection to etcd works.
func TestErrors(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to etcd could be established. Probably not running in a proper test environment.")
	}

	// Test with a bad MarshalFormat enum value

	client := createClient(t, etcd.MarshalFormat(19))
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
// Note: This test is only executed if the initial connection to etcd works.
func TestNil(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to etcd could be established. Probably not running in a proper test environment.")
	}

	// Test setting nil

	t.Run("set nil with JSON marshalling", func(t *testing.T) {
		client := createClient(t, etcd.JSON)
		err := client.Set("foo", nil)
		if err == nil {
			t.Error("Expected an error")
		}
	})

	t.Run("set nil with Gob marshalling", func(t *testing.T) {
		client := createClient(t, etcd.Gob)
		err := client.Set("foo", nil)
		if err == nil {
			t.Error("Expected an error")
		}
	})

	// Test passing nil or pointer to nil value for retrieval

	createTest := func(mf etcd.MarshalFormat) func(t *testing.T) {
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
	t.Run("get with nil / nil value parameter", createTest(etcd.JSON))
	t.Run("get with nil / nil value parameter", createTest(etcd.Gob))
}

// TestClose tests if the close method returns any errors.
//
// Note: This test is only executed if the initial connection to etcd works.
func TestClose(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to etcd could be established. Probably not running in a proper test environment.")
	}

	client := createClient(t, etcd.JSON)
	err := client.Close()
	if err != nil {
		t.Error(err)
	}
}

// checkConnection returns true if a connection could be made, false otherwise.
func checkConnection() bool {
	// The behaviour for New() seems to be inconsistent.
	// It should block at most for the specified time in DialTimeout.
	// In our case though New() doesn't block, but instead the following call does.
	// Maybe it's just the specific version we're using.
	// See https://github.com/etcd-io/etcd/issues/9829#issuecomment-438434795.
	// Use own timeout as workaround.
	// TODO: Remove workaround after etcd behaviour has been fixed or clarified.
	//cli, err := clientv3.NewFromURL("localhost:2379")
	config := clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 2 * time.Second,
	}
	okChan := make(chan bool, 1)
	go func() {
		cli, err := clientv3.New(config)
		if err != nil {
			log.Printf("An error occurred during testing the connection to the server: %v\n", err)
			okChan <- false
			return
		}
		statusRes, err := cli.Status(context.Background(), "localhost:2379")
		if err != nil {
			log.Printf("An error occurred during testing the connection to the server: %v\n", err)
			okChan <- false
			return
		} else if statusRes == nil {
			okChan <- false
			return
		}
		okChan <- true
	}()
	select {
	case <-okChan:
		return true
	case <-time.After(3 * time.Second):
		return false
	}
}

func createClient(t *testing.T, mf etcd.MarshalFormat) etcd.Client {
	options := etcd.DefaultOptions
	options.MarshalFormat = mf
	client, err := etcd.NewClient(options)
	if err != nil {
		t.Fatal(err)
	}
	return client
}
