package redis

import (
	"errors"

	"github.com/go-redis/redis"

	"github.com/philippgille/gokv/util"
)

// Client is a gokv.Store implementation for Redis.
type Client struct {
	c             *redis.Client
	marshalFormat MarshalFormat
}

// Set stores the given value for the given key.
// Values are automatically marshalled to JSON or gob (depending on the configuration).
// The key must not be "" and the value must not be nil.
func (c Client) Set(k string, v interface{}) error {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return err
	}

	// First turn the passed object into something that Redis can handle
	// (the Set method takes an interface{}, but the Get method only returns a string,
	// so it can be assumed that the interface{} parameter type is only for convenience
	// for a couple of builtin types like int etc.).
	var data []byte
	var err error
	switch c.marshalFormat {
	case JSON:
		data, err = util.ToJSON(v)
	case Gob:
		data, err = util.ToGob(v)
	default:
		err = errors.New("The store seems to be configured with a marshal format that's not implemented yet")
	}
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
// If no value is found it returns (false, nil).
// The key must not be "" and the pointer must not be nil.
func (c Client) Get(k string, v interface{}) (found bool, err error) {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return false, err
	}

	data, err := c.c.Get(k).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	}

	switch c.marshalFormat {
	case JSON:
		return true, util.FromJSON([]byte(data), v)
	case Gob:
		return true, util.FromGob([]byte(data), v)
	default:
		return true, errors.New("The store seems to be configured with a marshal format that's not implemented yet")
	}
}

// Delete deletes the stored value for the given key.
// Deleting a non-existing key-value pair does NOT lead to an error.
// The key must not be "".
func (c Client) Delete(k string) error {
	if err := util.CheckKey(k); err != nil {
		return err
	}

	_, err := c.c.Del(k).Result()
	return err
}

// Close closes the client.
// It must be called to release any open resources.
func (c Client) Close() error {
	return c.c.Close()
}

// MarshalFormat is an enum for the available (un-)marshal formats of this gokv.Store implementation.
type MarshalFormat int

const (
	// JSON is the MarshalFormat for (un-)marshalling to/from JSON
	JSON MarshalFormat = iota
	// Gob is the MarshalFormat for (un-)marshalling to/from gob
	Gob
)

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
	// (Un-)marshal format.
	// Optional (JSON by default).
	MarshalFormat MarshalFormat
}

// DefaultOptions is an Options object with default values.
// Address: "localhost:6379", Password: "", DB: 0, MarshalFormat: JSON
var DefaultOptions = Options{
	Address: "localhost:6379",
	// No need to set Password, DB or MarshalFormat
	// because their Go zero values are fine for that.
}

// NewClient creates a new Redis client.
func NewClient(options Options) (Client, error) {
	result := Client{}

	// Set default values
	if options.Address == "" {
		options.Address = DefaultOptions.Address
	}

	client := redis.NewClient(&redis.Options{
		Addr:     options.Address,
		Password: options.Password,
		DB:       options.DB,
	})

	err := client.Ping().Err()
	if err != nil {
		return result, err
	}

	result.c = client
	result.marshalFormat = options.MarshalFormat

	return result, nil
}
