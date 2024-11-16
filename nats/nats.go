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

type Client struct {
	kv    jetstream.KeyValue
	nc    *nats.Conn
	codec encoding.Codec
}

func (c Client) Set(k string, v any) error {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return err
	}

	data, err := c.codec.Marshal(v)
	if err != nil {
		return err
	}

	ctx := context.TODO() // Use TODO since context is not available yet.
	_, err = c.kv.Put(ctx, k, data)
	return err
}

func (c Client) Get(k string, v any) (found bool, err error) {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return false, err
	}

	ctx := context.TODO()
	entry, err := c.kv.Get(ctx, k)
	if err != nil {
		if errors.Is(err, jetstream.ErrKeyNotFound) {
			return false, nil
		}
		return false, err
	}

	return true, c.codec.Unmarshal(entry.Value(), v)
}

func (c Client) Delete(k string) error {
	if err := util.CheckKey(k); err != nil {
		return err
	}

	ctx := context.TODO()
	err := c.kv.Delete(ctx, k)
	if err != nil && errors.Is(err, jetstream.ErrKeyNotFound) {
		return nil
	}
	return err
}

func (c Client) Close() error {
	if c.nc != nil {
		c.nc.Close()
	}
	return nil
}

type Options struct {
	URL               string
	Bucket            string
	ConnectionTimeout *time.Duration
	Codec             encoding.Codec
}

var _defaultTimeout = 2 * time.Second

var DefaultOptions = Options{
	URL:               "nats://localhost:4222",
	ConnectionTimeout: &_defaultTimeout,
	Codec:             encoding.JSON,
}

func NewClient(options Options) (Client, error) {
	result := Client{}

	// Set default values
	if options.URL == "" {
		options.URL = DefaultOptions.URL
	}
	if options.ConnectionTimeout == nil {
		options.ConnectionTimeout = DefaultOptions.ConnectionTimeout
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
		// Try to create the bucket if it doesn't exist
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

	return result, nil
}

// Ensure that Client implements the gokv.Store interface.
var _ gokv.Store = (*Client)(nil)
