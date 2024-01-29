package hazelcast

import (
	"context"
	"fmt"

	hazelcast "github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/logger"

	"github.com/philippgille/gokv/encoding"
	"github.com/philippgille/gokv/util"
)

// Client is a gokv.Store implementation for Hazelcast.
type Client struct {
	c *hazelcast.Client
	// This map still works even after a temporary connection loss.
	m     *hazelcast.Map
	codec encoding.Codec
}

// Set stores the given value for the given key.
// Values are automatically marshalled to JSON or gob (depending on the configuration).
// The key must not be "" and the value must not be nil.
func (c Client) Set(k string, v any) error {
	ctx := context.Background()
	return c.SetWithContext(ctx, k, v)
}

// SetWithContext is exactly like Set function just with added context as first argument.
func (c Client) SetWithContext(ctx context.Context, k string, v any) error {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return err
	}

	// First turn the passed object into something that Hazelcast can handle
	data, err := c.codec.Marshal(v)
	if err != nil {
		return err
	}

	err = c.m.Set(ctx, k, data)
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
func (c Client) Get(k string, v any) (found bool, err error) {
	ctx := context.Background()
	return c.GetWithContext(ctx, k, v)
}

// GetWithContext is exactly like Get function just with added context as first argument.
func (c Client) GetWithContext(ctx context.Context, k string, v any) (found bool, err error) {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return false, err
	}

	hazelcastValue, err := c.m.Get(ctx, k)
	if err != nil {
		return false, err
	}
	// If no value was found return false
	if hazelcastValue == nil {
		return false, nil
	}
	data, ok := hazelcastValue.([]byte)
	if !ok {
		return false, fmt.Errorf("The returned value for key %v was expected to be a slice of bytes, but was type: %T", k, hazelcastValue)
	}

	return true, c.codec.Unmarshal(data, v)
}

// Delete deletes the stored value for the given key.
// The key must not be longer than 250 bytes (this is a restriction of Hazelcast).
// Deleting a non-existing key-value pair does NOT lead to an error.
// The key must not be "".
func (c Client) Delete(k string) error {
	ctx := context.Background()
	return c.DeleteWithContext(ctx, k)
}

// DeleteWithContext is exactly like Delete function just with added context as first argument.
func (c Client) DeleteWithContext(ctx context.Context, k string) error {
	if err := util.CheckKey(k); err != nil {
		return err
	}

	return c.m.Delete(ctx, k)
}

// Close closes the client.
// This must be called to properly shut down connections and services (e.g. HeartBeatService).
func (c Client) Close() error {
	c.c.Shutdown(context.Background())
	return nil
}

// Options are the options for the Hazelcast client.
type Options struct {
	// Address of one Hazelcast server, including port.
	// The client will delegate all operations to the given server.
	// If the server dies, the client will automatically switch to another server in the cluster.
	// Optional ("localhost:5701" by default).
	Address string
	// Name of the Hazelcast distributed map to use.
	// Optional ("gokv" by default).
	MapName string
	// Encoding format.
	// Optional (encoding.JSON by default).
	Codec encoding.Codec
}

// DefaultOptions is an Options object with default values.
// Addresses: "localhost:5701", MapName: "gokv", Codec: encoding.JSON
var DefaultOptions = Options{
	Address: "localhost:5701",
	MapName: "gokv",
	Codec:   encoding.JSON,
}

// NewClient creates a new Hazelcast client.
func NewClient(options Options) (Client, error) {
	result := Client{}

	// Set default values
	if options.Address == "" {
		options.Address = DefaultOptions.Address
	}
	if options.MapName == "" {
		options.MapName = DefaultOptions.MapName
	}
	if options.Codec == nil {
		options.Codec = DefaultOptions.Codec
	}

	config := hazelcast.NewConfig()
	config.Cluster.Network.SetAddresses(options.Address)
	config.Logger.Level = logger.OffLevel
	client, err := hazelcast.StartNewClientWithConfig(context.Background(), config)
	if err != nil {
		return result, err
	}

	hazelcastMap, err := client.GetMap(context.Background(), options.MapName)
	if err != nil {
		return result, err
	}

	result.c = client
	result.m = hazelcastMap
	result.codec = options.Codec

	return result, nil
}
