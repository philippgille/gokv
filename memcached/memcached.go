package memcached

import (
	"time"

	"github.com/bradfitz/gomemcache/memcache"

	"github.com/philippgille/gokv/encoding"
	"github.com/philippgille/gokv/util"
)

var defaultTimeout = 200 * time.Millisecond

// Client is a gokv.Store implementation for Memcached.
type Client struct {
	c     *memcache.Client
	codec encoding.Codec
}

// Set stores the given value for the given key.
// The key must not be longer than 250 bytes (this is a restriction of Memcached).
// Values are automatically marshalled to JSON or gob (depending on the configuration).
// The key must not be "" and the value must not be nil.
func (c Client) Set(k string, v any) error {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return err
	}

	// First turn the passed object into something that Memcached can handle
	data, err := c.codec.Marshal(v)
	if err != nil {
		return err
	}

	item := memcache.Item{
		Key:   k,
		Value: data,
	}
	err = c.c.Set(&item)
	if err != nil {
		return err
	}

	return nil
}

// Get retrieves the stored value for the given key.
// The key must not be longer than 250 bytes (this is a restriction of Memcached).
// You need to pass a pointer to the value, so in case of a struct
// the automatic unmarshalling can populate the fields of the object
// that v points to with the values of the retrieved object's values.
// If no value is found it returns (false, nil).
// The key must not be "" and the pointer must not be nil.
func (c Client) Get(k string, v any) (found bool, err error) {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return false, err
	}

	item, err := c.c.Get(k)
	// If no value was found return false
	if err == memcache.ErrCacheMiss {
		return false, nil
	} else if err != nil {
		return false, err
	}
	data := item.Value

	return true, c.codec.Unmarshal(data, v)
}

// Delete deletes the stored value for the given key.
// The key must not be longer than 250 bytes (this is a restriction of Memcached).
// Deleting a non-existing key-value pair does NOT lead to an error.
// The key must not be "".
func (c Client) Delete(k string) error {
	if err := util.CheckKey(k); err != nil {
		return err
	}

	err := c.c.Delete(k)
	if err == memcache.ErrCacheMiss {
		return nil
	}
	return err
}

// Close closes the client.
// In the Memcached implementation this doesn't have any effect.
func (c Client) Close() error {
	return nil
}

// Options are the options for the Memcached client.
type Options struct {
	// Addresses of all Memcached servers, including their port.
	// If a server is listed multiple times it gets a proportional amount of weight.
	// Optional ("localhost:11211" by default).
	Addresses []string
	// Timeout for requests.
	// The gomemcache package uses a default of 100 milliseconds,
	// which seems ok for the use of a caching server, but too low for the use of an (albeit ephemeral) key-value storage.
	// Optional (200 milliseconds by default).
	Timeout *time.Duration
	// Maximum number of idle connections per Memcached server.
	// Default max connections on the server are 1024, so 100 from one client should be fine.
	// The gomemcache package uses a default of 2, which seems to be too low regarding its description:
	// "This should be set to a number higher than your peak parallel requests".
	// 0 will lead to the default value being used.
	// Optional (100 by default).
	MaxIdleConns int
	// Encoding format.
	// Optional (encoding.JSON by default).
	Codec encoding.Codec
}

// DefaultOptions is an Options object with default values.
// Addresses: "localhost:11211", Timeout: 200 milliseconds, MaxIdleConns: 100, Codec: encoding.JSON
var DefaultOptions = Options{
	Addresses:    []string{"localhost:11211"},
	Timeout:      &defaultTimeout,
	MaxIdleConns: 100,
	Codec:        encoding.JSON,
}

// NewClient creates a new Memcached client.
func NewClient(options Options) (Client, error) {
	result := Client{}

	// Set default values
	if len(options.Addresses) == 0 {
		options.Addresses = DefaultOptions.Addresses
	}
	if options.Timeout == nil {
		options.Timeout = DefaultOptions.Timeout
	}
	if options.MaxIdleConns == 0 {
		options.MaxIdleConns = DefaultOptions.MaxIdleConns
	}
	if options.Codec == nil {
		options.Codec = DefaultOptions.Codec
	}

	mc := memcache.New(options.Addresses...)
	mc.Timeout = *options.Timeout
	mc.MaxIdleConns = options.MaxIdleConns

	result.c = mc
	result.codec = options.Codec

	return result, nil
}
