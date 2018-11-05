package redis

import (
	"github.com/go-redis/redis"

	"github.com/philippgille/gokv/util"
)

// Client is a gokv.Store implementation for Redis.
type Client struct {
	c *redis.Client
}

// Set stores the given object for the given key.
// Values are marshalled to JSON automatically.
func (c Client) Set(k string, v interface{}) error {
	// First turn the passed object into something that Redis can handle
	// (the Set method takes an interface{}, but the Get method only returns a string,
	// so it can be assumed that the interface{} parameter type is only for convenience
	// for a couple of builtin types like int etc.).
	data, err := util.ToJSON(v)
	if err != nil {
		return err
	}

	err = c.c.Set(k, string(data), 0).Err()
	if err != nil {
		return err
	}
	return nil
}

// Get retrieves the stored value for the given key.
// You need to pass a pointer to the value, so in case of a struct
// the automatic unmarshalling can populate the fields of the object
// that v points to with the values of the retrieved object's values.
func (c Client) Get(k string, v interface{}) (bool, error) {
	data, err := c.c.Get(k).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	}

	return true, util.FromJSON([]byte(data), v)
}

// Delete deletes the stored value for the given key.
// Deleting a non-existing key-value pair does NOT lead to an error.
func (c Client) Delete(k string) error {
	_, err := c.c.Del(k).Result()
	return err
}

// Options are the options for the Redis client.
type Options struct {
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

// DefaultOptions is an Options object with default values.
// Address: "localhost:6379", Password: "", DB: 0
var DefaultOptions = Options{
	Address: "localhost:6379",
	// No need to set Password or DB, since their Go zero values are fine for that
}

// NewClient creates a new Redis client.
func NewClient(options Options) Client {
	// Set default values
	if options.Address == "" {
		options.Address = DefaultOptions.Address
	}
	return Client{
		c: redis.NewClient(&redis.Options{
			Addr:     options.Address,
			Password: options.Password,
			DB:       options.DB,
		}),
	}
}
