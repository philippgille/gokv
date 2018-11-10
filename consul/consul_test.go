package consul_test

import (
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/consul/api"

	"github.com/philippgille/gokv/consul"
	"github.com/philippgille/gokv/test"
)

// TestClient tests if reading from, writing to and deleting from the store works properly.
// A struct is used as value. See TestTypes() for a test that is simpler but tests all types.
//
// Note: This test is only executed if the initial connection to Consul works.
func TestClient(t *testing.T) {
	if !checkConsulConnection() {
		t.Skip("No connection to Consul could be established. Probably not running in a proper test environment.")
	}

	// Test with JSON
	t.Run("JSON", func(t *testing.T) {
		client := createClient(t, consul.JSON)
		test.TestStore(client, t)
	})

	// Test with gob
	t.Run("gob", func(t *testing.T) {
		client := createClient(t, consul.Gob)
		test.TestStore(client, t)
	})
}

// TestTypes tests if setting and getting values works with all Go types.
//
// Note: This test is only executed if the initial connection to Consul works.
func TestTypes(t *testing.T) {
	if !checkConsulConnection() {
		t.Skip("No connection to Consul could be established. Probably not running in a proper test environment.")
	}

	// Test with JSON
	t.Run("JSON", func(t *testing.T) {
		client := createClient(t, consul.JSON)
		test.TestTypes(client, t)
	})

	// Test with gob
	t.Run("gob", func(t *testing.T) {
		client := createClient(t, consul.Gob)
		test.TestTypes(client, t)
	})
}

// TestClientConcurrent launches a bunch of goroutines that concurrently work with the Consul client.
//
// Note: This test is only executed if the initial connection to Consul works.
func TestClientConcurrent(t *testing.T) {
	if !checkConsulConnection() {
		t.Skip("No connection to Consul could be established. Probably not running in a proper test environment.")
	}

	client := createClient(t, consul.JSON)

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

// checkConsulConnection returns true if a connection could be made, false otherwise.
func checkConsulConnection() bool {
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		return false
	}
	res, err := client.Status().Leader()
	if err != nil || res == "" {
		return false
	}
	return true
}

func createClient(t *testing.T, mf consul.MarshalFormat) consul.Client {
	options := consul.DefaultOptions
	options.Folder = "test_" + strconv.FormatInt(time.Now().Unix(), 10)
	options.MarshalFormat = mf
	client, err := consul.NewClient(options)
	if err != nil {
		t.Error(err)
	}
	return client
}
