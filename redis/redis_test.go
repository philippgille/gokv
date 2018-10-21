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

// TestRedisClient tests if reading and writing to the store works properly.
//
// Note: This test is only executed if the initial connection to Redis works.
func TestClient(t *testing.T) {
	if !checkRedisConnection(testDbNumber) {
		t.Skip("No connection to Redis could be established. Probably not running in a proper test environment.")
	}

	deleteRedisDb(testDbNumber) // Prep for previous test runs
	options := redis.Options{
		DB: testDbNumber,
	}
	client := redis.NewClient(options)

	test.TestStore(client, t)
}

// TestRedisClientConcurrent launches a bunch of goroutines that concurrently work with the Redis client.
func TestClientConcurrent(t *testing.T) {
	if !checkRedisConnection(testDbNumber) {
		t.Skip("No connection to Redis could be established. Probably not running in a proper test environment.")
	}

	deleteRedisDb(testDbNumber) // Prep for previous test runs
	options := redis.Options{
		DB: testDbNumber,
	}
	client := redis.NewClient(options)

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
			t.Errorf("No value was found, but should have been")
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
