package gokv

import (
	"github.com/go-redis/redis"
)

// RedisClient is a StorageClient implementation for Redis.
type RedisClient struct {
	c *redis.Client
}

// Set stores the given object for the given key.
func (c RedisClient) Set(k string, v interface{}) error {
	// First turn the passed object into something that Redis can handle
	// (the Set method takes an interface{}, but the Get method only returns a string,
	// so it can be assumed that the interface{} parameter type is only for convenience
	// for a couple of builtin types like int etc.).
	data, err := toJSON(v)
	if err != nil {
		return err
	}

	err = c.c.Set(k, string(data), 0).Err()
	if err != nil {
		return err
	}
	return nil
}

// Get retrieves the object for the given key and points the passed pointer to it.
func (c RedisClient) Get(k string, v interface{}) (bool, error) {
	data, err := c.c.Get(k).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	}

	return true, fromJSON([]byte(data), v)
}

// RedisOptions are the options for the Redis DB.
type RedisOptions struct {
	// Address of the Redis server, including the port.
	// Optional ("localhost:6379" by default).
	Address string
	// Password for the Redis server.
	// Optional ("" by default).
	Password string
	// DB to use.
	// Optional (0 by default).
	DB int
}

// DefaultRedisOptions is a RedisOptions object with default values.
// Address: "localhost:6379", Password: "", DB: 0
var DefaultRedisOptions = RedisOptions{
	Address: "localhost:6379",
	// No need to set Password or DB, since their Go zero values are fine for that
}

// NewRedisClient creates a new RedisClient.
func NewRedisClient(redisOptions RedisOptions) RedisClient {
	// Set default values
	if redisOptions.Address == "" {
		redisOptions.Address = DefaultRedisOptions.Address
	}
	return RedisClient{
		c: redis.NewClient(&redis.Options{
			Addr:     redisOptions.Address,
			Password: redisOptions.Password,
			DB:       redisOptions.DB,
		}),
	}
}
