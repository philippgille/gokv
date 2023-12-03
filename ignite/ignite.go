package ignite

import (
	"fmt"
	"net"
	"time"

	ignite "github.com/amsokol/ignite-go-client/binary/v1"

	"github.com/philippgille/gokv/encoding"
	"github.com/philippgille/gokv/util"
)

// Client is a gokv.Store implementation for Apache Ignite.
type Client struct {
	c         ignite.Client
	cacheName string
	codec     encoding.Codec
}

// Set stores the given value for the given key.
// Values are automatically marshalled to JSON or gob (depending on the configuration).
// The key must not be "" and the value must not be nil.
func (c Client) Set(k string, v any) error {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return err
	}

	// First turn the passed object into something that Apache Ignite can handle
	data, err := c.codec.Marshal(v)
	if err != nil {
		return err
	}

	return c.c.CachePut(c.cacheName, true, k, data)
}

// Get retrieves the stored value for the given key.
// You need to pass a pointer to the value, so in case of a struct
// the automatic unmarshalling can populate the fields of the object
// that v points to with the values of the retrieved object's values.
// If no value is found it returns (false, nil).
// The key must not be "" and the pointer must not be nil.
func (c Client) Get(k string, v any) (found bool, err error) {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return false, err
	}

	dataIface, err := c.c.CacheGet(c.cacheName, true, k)
	if err != nil {
		return false, err
	}
	// If no value was found return false.
	// Due to the used package we can't differentiate between a nil value and a value that's not found.
	// But nil values can't be set with both the used Go package as well as the official .NET Core thin client
	// (when using `string` as value type), so maybe nil values aren't allowed by Ignite anyway.
	if dataIface == nil {
		return false, nil
	}
	data, ok := dataIface.([]byte)
	if !ok {
		return true, fmt.Errorf("The value for key %v is expected to be a slice of bytes, but its type is: %T", k, dataIface)
	}

	return true, c.codec.Unmarshal(data, v)
}

// Delete deletes the stored value for the given key.
// Deleting a non-existing key-value pair does NOT lead to an error.
// The key must not be "".
func (c Client) Delete(k string) error {
	if err := util.CheckKey(k); err != nil {
		return err
	}

	_, err := c.c.CacheRemoveKey(c.cacheName, false, k)
	if err != nil {
		return err
	}
	return err
}

// Close closes the client.
// It must be called to shut down all connections to the Apache Ignite server.
func (c Client) Close() error {
	return c.c.Close()
}

// Options are the options for the Apache Ignite client.
type Options struct {
	// Server address without port.
	// Optional ("localhost" by default).
	Host string
	// Server binary protocol connector port.
	// See https://apacheignite.readme.io/docs/binary-client-protocol#section-tcp-socket.
	// Optional (10800 by default).
	Port int
	// Username.
	// Optional ("" by default).
	Username string
	// Password.
	// Optional ("" by default).
	Password string
	// Name of the cache.
	// Optional ("gokv" by default).
	CacheName string
	// Encoding format.
	// Optional (encoding.JSON by default).
	Codec encoding.Codec
}

// DefaultOptions is an Options object with default values.
// Host: "localhost", Port: 10800, CacheName: "gokv", Codec: encoding.JSON
var DefaultOptions = Options{
	Host:      "localhost",
	Port:      10800,
	CacheName: "gokv",
	Codec:     encoding.JSON,
}

// NewClient creates a new Apache Ignite client.
//
// You must call the Close() method on the client when you're done working with it.
func NewClient(options Options) (Client, error) {
	result := Client{}

	// Set default values
	if options.Host == "nil" {
		options.Host = DefaultOptions.Host
	}
	if options.Port == 0 {
		options.Port = DefaultOptions.Port
	}
	if options.CacheName == "" {
		options.CacheName = DefaultOptions.CacheName
	}
	if options.Codec == nil {
		options.Codec = DefaultOptions.Codec
	}

	connInfo := ignite.ConnInfo{
		// This timeout should just be for the initial dialing,
		// not for subsequent requests.
		Dialer: net.Dialer{
			Timeout: 2 * time.Second,
		},
		Host:     options.Host,
		Major:    1,
		Minor:    1,
		Network:  "tcp",
		Password: options.Password,
		Port:     options.Port,
		Username: options.Username,
		// Go zero values for Patch and TLSConfig.
	}
	c, err := ignite.Connect(connInfo)
	if err != nil {
		return result, err
	}

	// Create cache if it doesn't exist yet.
	err = c.CacheGetOrCreateWithName(options.CacheName)
	if err != nil {
		return result, err
	}

	result.c = c
	result.cacheName = options.CacheName
	result.codec = options.Codec

	return result, nil
}
