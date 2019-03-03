package hazelcast

import (
	"fmt"
	"time"

	hazelcast "github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/config/property"
	"github.com/hazelcast/hazelcast-go-client/core"
	"github.com/hazelcast/hazelcast-go-client/core/logger"

	"github.com/philippgille/gokv/encoding"
	"github.com/philippgille/gokv/util"
)

var defaultTimeout = 200 * time.Millisecond

// Client is a gokv.Store implementation for Hazelcast.
type Client struct {
	// hazelcast.Client is an interface, so don't use a pointer here
	// For the Set(), Get() and Delete() operations we only need the core.Map,
	// but we need the client for Close().
	c hazelcast.Client
	// core.Map is an interface, so don't use a pointer here
	//
	// TODO: When a Hazelcast server dies and the client creates new connections to the new server,
	// does the map still work or do we need to get the map from the client again?
	m     core.Map
	codec encoding.Codec
}

// Set stores the given value for the given key.
// Values are automatically marshalled to JSON or gob (depending on the configuration).
// The key must not be "" and the value must not be nil.
func (c Client) Set(k string, v interface{}) error {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return err
	}

	// First turn the passed object into something that Hazelcast can handle
	data, err := c.codec.Marshal(v)
	if err != nil {
		return err
	}

	err = c.m.Set(k, data)
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

	hazelcastValue, err := c.m.Get(k)
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
	if err := util.CheckKey(k); err != nil {
		return err
	}

	return c.m.Delete(k)
}

// Close closes the client.
// This must be called to properly shut down connections and services (e.g. HeartBeatService).
func (c Client) Close() error {
	c.c.Shutdown()
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
// Addresses: "localhost:11211", Timeout: 200 milliseconds, MaxIdleConns: 100, Codec: encoding.JSON
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
	config.NetworkConfig().AddAddress(options.Address)
	// During NewClientWithConfig() Hazelcast calls internal.init() on the client,
	// which in turn calls internal.initLogger(), which just sets the logger.DefaultLogger if no logger is defined.
	// We don't want the verbose logging, so we need to turn this off.
	// It doesn't work with setting a logger in the following way:
	//hazelcastDefaultLogger := logger.New()
	//offLogLevel, _ := logger.GetLogLevel(logger.OffLevel)
	//hazelcastDefaultLogger.Level = offLogLevel
	//config.LoggerConfig().SetLogger(hazelcastDefaultLogger)
	// Instead, the correct way is documented in the GitHub repository's README:
	// https://github.com/hazelcast/hazelcast-go-client#782-logging-configuration
	config.SetProperty(property.LoggingLevel.Name(), logger.OffLevel)
	client, err := hazelcast.NewClientWithConfig(config)
	if err != nil {
		return result, err
	}

	hazelcastMap, err := client.GetMap(options.MapName)
	if err != nil {
		return result, err
	}

	result.c = client
	result.m = hazelcastMap
	result.codec = options.Codec

	return result, nil
}
