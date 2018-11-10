package redis_test

import (
	"log"
	"strconv"
	"sync"
	"testing"

	goredis "github.com/go-redis/redis"

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
	if !checkRedisConnection(testDbNumber) {
		t.Skip("No connection to Redis could be established. Probably not running in a proper test environment.")
	}
	deleteRedisDb(testDbNumber) // Prep for previous test runs

	// Test with JSON
	t.Run("JSON", func(t *testing.T) {
		client := createClient(t, redis.JSON)
		test.TestStore(client, t)
	})

	// Test with gob
	t.Run("gob", func(t *testing.T) {
		client := createClient(t, redis.Gob)
		test.TestStore(client, t)
	})
}

// TestTypes tests if setting and getting values works with all Go types.
//
// Note: This test is only executed if the initial connection to Redis works.
func TestTypes(t *testing.T) {
	if !checkRedisConnection(testDbNumber) {
		t.Skip("No connection to Redis could be established. Probably not running in a proper test environment.")
	}
	deleteRedisDb(testDbNumber) // Prep for previous test runs

	// Test with JSON
	t.Run("JSON", func(t *testing.T) {
		client := createClient(t, redis.JSON)
		test.TestTypes(client, t)
	})

	// Test with gob
	t.Run("gob", func(t *testing.T) {
		client := createClient(t, redis.Gob)
		test.TestTypes(client, t)
	})
}

// TestClientConcurrent launches a bunch of goroutines that concurrently work with the Redis client.
//
// Note: This test is only executed if the initial connection to Redis works.
func TestClientConcurrent(t *testing.T) {
	if !checkRedisConnection(testDbNumber) {
		t.Skip("No connection to Redis could be established. Probably not running in a proper test environment.")
	}
	deleteRedisDb(testDbNumber) // Prep for previous test runs

	client := createClient(t, redis.JSON)

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

// checkRedisConnection returns true if a connection could be made, false otherwise.
func checkRedisConnection(number int) bool {
	client := goredis.NewClient(&goredis.Options{
		Addr:     redis.DefaultOptions.Address,
		Password: redis.DefaultOptions.Password,
		DB:       number,
	})
	err := client.Ping().Err()
	if err != nil {
		log.Printf("An error occurred during testing the connection to Redis: %v\n", err)
		return false
	}
	return true
}

// deleteRedisDb deletes all entries of the given DB
func deleteRedisDb(number int) error {
	client := goredis.NewClient(&goredis.Options{
		Addr:     redis.DefaultOptions.Address,
		Password: redis.DefaultOptions.Password,
		DB:       number,
	})
	return client.FlushDB().Err()
}

func createClient(t *testing.T, mf redis.MarshalFormat) redis.Client {
	options := redis.Options{
		DB:            testDbNumber,
		MarshalFormat: mf,
	}
	client := redis.NewClient(options)
	return client
}
