package gokv_test

import (
	"log"
	"strconv"
	"sync"
	"testing"

	"github.com/go-redis/redis"

	"github.com/philippgille/ln-paywall/ln"
	"github.com/philippgille/ln-paywall/storage"
	"github.com/philippgille/ln-paywall/wall"
)

// Don't use the default number ("0"),
// which could lead to valuable data being deleted when a developer accidentally runs the test with valuable data in DB 0.
var testDbNumber = 15 // 16 DBs by default (unchanged config), starting with 0

// TestRedisClientImpl tests if the RedisClient struct implements the StorageClient interface.
// This doesn't happen at runtime, but at compile time.
func TestRedisClientImpl(t *testing.T) {
	t.SkipNow()
	invoiceOptions := wall.InvoiceOptions{}
	lnClient := ln.LNDclient{}
	redisClient := storage.RedisClient{}
	wall.NewHandlerFuncMiddleware(invoiceOptions, lnClient, redisClient)
	wall.NewHandlerMiddleware(invoiceOptions, lnClient, redisClient)
	wall.NewGinMiddleware(invoiceOptions, lnClient, redisClient)
}

// TestRedisClient tests if reading and writing to the storage works properly.
//
// Note: This test is only executed if the initial connection to Redis works.
func TestRedisClient(t *testing.T) {
	if !checkRedisConnection(testDbNumber) {
		t.Skip("No connection to Redis could be established. Probably not running in a proper test environment.")
	}

	deleteRedisDb(testDbNumber) // Prep for previous test runs
	redisOptions := storage.RedisOptions{
		DB: testDbNumber,
	}
	redisClient := storage.NewRedisClient(redisOptions)

	testStorageClient(redisClient, t)
}

// TestRedisClientConcurrent launches a bunch of goroutines that concurrently work with the Redis client.
func TestRedisClientConcurrent(t *testing.T) {
	if !checkRedisConnection(testDbNumber) {
		t.Skip("No connection to Redis could be established. Probably not running in a proper test environment.")
	}

	deleteRedisDb(testDbNumber) // Prep for previous test runs
	redisOptions := storage.RedisOptions{
		DB: testDbNumber,
	}
	redisClient := storage.NewRedisClient(redisOptions)

	goroutineCount := 1000

	waitGroup := sync.WaitGroup{}
	waitGroup.Add(goroutineCount) // Must be called before any goroutine is started
	for i := 0; i < goroutineCount; i++ {
		go interactWithStorage(redisClient, strconv.Itoa(i), t, &waitGroup)
	}
	waitGroup.Wait()

	// Now make sure that all values are in the storage
	expected := foo{}
	for i := 0; i < goroutineCount; i++ {
		actualPtr := new(foo)
		found, err := redisClient.Get(strconv.Itoa(i), actualPtr)
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
	redisClient := redis.NewClient(&redis.Options{
		Addr:     storage.DefaultRedisOptions.Address,
		Password: storage.DefaultRedisOptions.Password,
		DB:       number,
	})
	err := redisClient.Ping().Err()
	if err != nil {
		log.Printf("An error occurred during testing the connection to Redis: %v\n", err)
		return false
	}
	return true
}

// deleteRedisDb deletes all entries of the given DB
func deleteRedisDb(number int) error {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     storage.DefaultRedisOptions.Address,
		Password: storage.DefaultRedisOptions.Password,
		DB:       number,
	})
	return redisClient.FlushDB().Err()
}
