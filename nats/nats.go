package nats

import (
	"context"
	"errors"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/philippgille/gokv"
	"github.com/philippgille/gokv/encoding"
	"github.com/philippgille/gokv/util"
)

// Client is a gokv.Store implementation for NATS JetStream KV.
type Client struct {
	kv      jetstream.KeyValue
	nc      *nats.Conn
	codec   encoding.Codec
	timeout time.Duration
}

// Set stores the given value for the given key.
// Values are automatically marshalled to JSON or gob (depending on the configuration).
// The key must not be "" and the value must not be nil.
func (c Client) Set(k string, v any) error {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return err
	}

	data, err := c.codec.Marshal(v)
	if err != nil {
		return err
	}

	ctxWithTimeout, cancel := context.WithTimeout(context.TODO(), c.timeout) // Use TODO since context is not available yet.
	defer cancel()
	_, err = c.kv.Put(ctxWithTimeout, k, data)
	return err
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

	ctxWithTimeout, cancel := context.WithTimeout(context.TODO(), c.timeout)
	defer cancel()
	entry, err := c.kv.Get(ctxWithTimeout, k)
	if err != nil {
		if errors.Is(err, jetstream.ErrKeyNotFound) {
			return false, nil
		}
		return false, err
	}

	return true, c.codec.Unmarshal(entry.Value(), v)
}

// Delete deletes the stored value for the given key.
// Deleting a non-existing key-value pair does NOT lead to an error.
// The key must not be "".
func (c Client) Delete(k string) error {
	if err := util.CheckKey(k); err != nil {
		return err
	}

	ctxWithTimeout, cancel := context.WithTimeout(context.TODO(), c.timeout)
	defer cancel()
	err := c.kv.Delete(ctxWithTimeout, k)
	if err != nil && errors.Is(err, jetstream.ErrKeyNotFound) {
		return nil
	}
	return err
}

// Close closes the client.
// It must be called to release resources used by the NATS connection.
func (c Client) Close() error {
	if c.nc != nil {
		return c.nc.Drain()
	}
	return nil
}

// Options are the options for the NATS client.
type Options struct {
	// URL is the NATS server URL.
	// Optional ("nats://localhost:4222" by default).
	URL string
	// Bucket is the name of the KV bucket to use.
	// If the bucket doesn't exist, it will be created.
	// To follow a best practice, always create a bucket in advance on the server side.
	// Required.
	Bucket string
	// Connection timeout.
	// Optional (2 seconds by default).
	ConnectionTimeout *time.Duration
	// Operation timeout.
	// Optional (2 seconds by default).
	OperationTimeout *time.Duration
	// Encoding format.
	// Optional (encoding.JSON by default).
	Codec encoding.Codec
}

var _defaultTimeout = 2 * time.Second

// DefaultOptions is an Options object with default values.
// URL: "nats://localhost:4222", ConnectionTimeout: 2 * time.Second, Codec: encoding.JSON
// Note: Bucket is required and must be set by the user.
var DefaultOptions = Options{
	URL:               "nats://localhost:4222",
	ConnectionTimeout: &_defaultTimeout,
	OperationTimeout:  &_defaultTimeout,
	Codec:             encoding.JSON,
}

// NewClient creates a new NATS client.
func NewClient(options Options) (Client, error) {
	result := Client{}

	// Set default values
	if options.URL == "" {
		options.URL = DefaultOptions.URL
	}
	if options.ConnectionTimeout == nil {
		options.ConnectionTimeout = DefaultOptions.ConnectionTimeout
	}
	if options.OperationTimeout == nil {
		options.OperationTimeout = DefaultOptions.OperationTimeout
	}
	if options.Codec == nil {
		options.Codec = DefaultOptions.Codec
	}
	if options.Bucket == "" {
		return result, errors.New("bucket name is required")
	}

	// Connect to NATS
	nc, err := nats.Connect(options.URL, nats.Timeout(*options.ConnectionTimeout))
	if err != nil {
		return result, err
	}

	// Create or get the Key-Value store
	js, err := jetstream.New(nc)
	if err != nil {
		return result, err
	}

	ctx := context.TODO()
	kv, err := js.KeyValue(ctx, options.Bucket)
	if err != nil {
		if !errors.Is(err, jetstream.ErrBucketNotFound) {
			return result, err
		}
		// Try to create the bucket if it doesn't exist.
		// It doesn't handle the concurrent creation of the bucket by multiple clients.
		// CreateOrUpdateKeyValue() might be a solution, but it's not a good practice.
		// The bucket should be created in advance on the server side, or in a dedicated process.
		kv, err = js.CreateKeyValue(ctx, jetstream.KeyValueConfig{
			Bucket: options.Bucket,
		})
		if err != nil {
			return result, err
		}
	}

	result.kv = kv
	result.codec = options.Codec
	// Store the connection in the client,
	// so it can be closed when Close() is called.
	result.nc = nc
	result.timeout = *options.OperationTimeout

	return result, nil
}

// Ensure that Client implements the gokv.Store interface.
var _ gokv.Store = (*Client)(nil)
