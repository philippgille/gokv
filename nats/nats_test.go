package nats_test

import (
	"log"
	"testing"
	"time"

	"github.com/nats-io/nats-server/v2/server"

	"github.com/philippgille/gokv/encoding"
	"github.com/philippgille/gokv/nats"
	"github.com/philippgille/gokv/test"
)

// Test server instance for all tests
var ts *server.Server

// TestMain initializes the embedded nats server and runs the tests.
func TestMain(m *testing.M) {
	// Start embedded NATS server
	opts := &server.Options{
		Host:      "127.0.0.1",
		Port:      4222,
		JetStream: true,
		// Reduce logging noise in tests
		Debug: false,
		Trace: false,
	}

	var err error
	ts, err = server.NewServer(opts)
	if err != nil {
		log.Fatal(err)
	}

	go ts.Start()
	// Wait for server to be ready
	if !ts.ReadyForConnections(4 * time.Second) {
		log.Fatal("Unable to start embedded NATS server")
	}

	// Run tests
	code := m.Run()

	// Shutdown server
	ts.Shutdown()
	ts.WaitForShutdown()

	// Exit with test result code
	log.Printf("Exiting with code: %d", code)
}

// TestClient tests if reading from, writing to and deleting from the store works properly.
// A struct is used as value. See TestTypes() for a test that is simpler but tests all types.
func TestClient(t *testing.T) {
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
func TestTypes(t *testing.T) {
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

// TestClientConcurrent launches a bunch of goroutines that concurrently work with the NATS client.
func TestClientConcurrent(t *testing.T) {
	client := createClient(t, encoding.JSON)
	defer client.Close()

	goroutineCount := 1000

	test.TestConcurrentInteractions(t, goroutineCount, client)
}

// TestErrors tests some error cases.
func TestErrors(t *testing.T) {
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

	// Test client creation with bad options
	options := nats.Options{
		URL: "nats://nonexistent:4222",
	}
	_, err = nats.NewClient(options)
	if err == nil {
		t.Error("Expected an error for connection to non-existent server")
	}

	options = nats.Options{
		URL:    "nats://localhost:4222",
		Bucket: "",
	}
	_, err = nats.NewClient(options)
	if err == nil || err.Error() != "bucket name is required" {
		t.Error("Expected an error for empty bucket name")
	}
}

// TestNil tests the behaviour when passing nil or pointers to nil values to some methods.
func TestNil(t *testing.T) {
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
func TestClose(t *testing.T) {
	client := createClient(t, encoding.JSON)
	err := client.Close()
	if err != nil {
		t.Error(err)
	}
}

// TestDefaultTimeout tests if the client works with the default timeout.
func TestDefaultTimeout(t *testing.T) {
	options := nats.Options{
		Bucket: "test-default-timeout",
	}
	client, err := nats.NewClient(options)
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
		t.Errorf("Expected %v, but was %v", "bar", *vPtr)
	}
}

func createClient(t *testing.T, codec encoding.Codec) nats.Client {
	if ts == nil || !ts.Running() {
		t.Fatal("Test server not running")
	}

	connectionTimeout := 2 * time.Second
	operationTimeout := 1 * time.Second
	options := nats.Options{
		URL:               ts.ClientURL(),
		Bucket:            "test-bucket",
		ConnectionTimeout: &connectionTimeout,
		OperationTimeout:  &operationTimeout,
		Codec:             codec,
	}
	client, err := nats.NewClient(options)
	if err != nil {
		t.Fatal(err)
	}
	return client
}
