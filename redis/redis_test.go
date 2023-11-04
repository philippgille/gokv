package redis_test

import (
	"context"
	"log"
	"testing"

	goredis "github.com/redis/go-redis/v9"

	"github.com/philippgille/gokv/encoding"
	"github.com/philippgille/gokv/redis"
	"github.com/philippgille/gokv/test"
)

// Don't use the default number ("0"),
// which could lead to valuable data being deleted when a developer accidentally runs the test with valuable data in DB 0.
var testDbNumber = 15 // 16 DBs by default (unchanged config), starting with 0

// TestClient tests if reading from, writing to and deleting from the store works properly.
// A struct is used as value. See TestTypes() for a test that is simpler but tests all types.
//
// Note: This test is only executed if the initial connection to Redis works.
func TestClient(t *testing.T) {
	if !checkConnection(testDbNumber) {
		t.Skip("No connection to Redis could be established. Probably not running in a proper test environment.")
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
// Note: This test is only executed if the initial connection to Redis works.
func TestTypes(t *testing.T) {
	if !checkConnection(testDbNumber) {
		t.Skip("No connection to Redis could be established. Probably not running in a proper test environment.")
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

// TestClientConcurrent launches a bunch of goroutines that concurrently work with the Redis client.
//
// Note: This test is only executed if the initial connection to Redis works.
func TestClientConcurrent(t *testing.T) {
	if !checkConnection(testDbNumber) {
		t.Skip("No connection to Redis could be established. Probably not running in a proper test environment.")
	}

	client := createClient(t, encoding.JSON)
	defer client.Close()

	goroutineCount := 1000

	test.TestConcurrentInteractions(t, goroutineCount, client)
}

// TestErrors tests some error cases.
//
// Note: This test is only executed if the initial connection to Redis works.
func TestErrors(t *testing.T) {
	if !checkConnection(testDbNumber) {
		t.Skip("No connection to Redis could be established. Probably not running in a proper test environment.")
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
// Note: This test is only executed if the initial connection to Redis works.
func TestNil(t *testing.T) {
	if !checkConnection(testDbNumber) {
		t.Skip("No connection to Redis could be established. Probably not running in a proper test environment.")
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
// Note: This test is only executed if the initial connection to Redis works.
func TestClose(t *testing.T) {
	if !checkConnection(testDbNumber) {
		t.Skip("No connection to Redis could be established. Probably not running in a proper test environment.")
	}

	client := createClient(t, encoding.JSON)
	err := client.Close()
	if err != nil {
		t.Error(err)
	}
}

// checkConnection returns true if a connection could be made, false otherwise.
func checkConnection(number int) bool {
	client := goredis.NewClient(&goredis.Options{
		Addr:     redis.DefaultOptions.Address,
		Password: redis.DefaultOptions.Password,
		DB:       number,
	})
	defer client.Close()
	err := client.Ping(context.Background()).Err()
	if err != nil {
		log.Printf("An error occurred during testing the connection to the server: %v\n", err)
		return false
	}
	return true
}

func createClient(t *testing.T, codec encoding.Codec) redis.Client {
	options := redis.Options{
		DB:    testDbNumber,
		Codec: codec,
	}
	client, err := redis.NewClient(options)
	if err != nil {
		t.Fatal(err)
	}
	return client
}
