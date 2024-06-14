package etcd

import (
	"context"
	"errors"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"

	"github.com/philippgille/gokv/encoding"
	"github.com/philippgille/gokv/util"
)

var defaultTimeout = 200 * time.Millisecond

// Client is a gokv.Store implementation for etcd.
type Client struct {
	c       *clientv3.Client
	timeOut time.Duration
	codec   encoding.Codec
}

// Gets underlying store to allow user manipulate object directly.
func (c Client) GetStore() *clientv3.Client {
	return c.c
}

// Set stores the given value for the given key.
// Values are automatically marshalled to JSON or gob (depending on the configuration).
// The key must not be "" and the value must not be nil.
func (c Client) Set(k string, v any) error {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return err
	}

	// First turn the passed object into something that etcd can handle
	data, err := c.codec.Marshal(v)
	if err != nil {
		return err
	}

	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), c.timeOut)
	defer cancel()
	_, err = c.c.Put(ctxWithTimeout, k, string(data))
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
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return false, err
	}

	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), c.timeOut)
	defer cancel()
	getRes, err := c.c.Get(ctxWithTimeout, k)
	if err != nil {
		return false, err
	}
	kvs := getRes.Kvs
	// If no value was found return false
	if len(kvs) == 0 {
		return false, nil
	}
	data := kvs[0].Value

	return true, c.codec.Unmarshal(data, v)
}

// Delete deletes the stored value for the given key.
// Deleting a non-existing key-value pair does NOT lead to an error.
// The key must not be "".
func (c Client) Delete(k string) error {
	if err := util.CheckKey(k); err != nil {
		return err
	}

	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), c.timeOut)
	defer cancel()
	_, err := c.c.Delete(ctxWithTimeout, k)
	return err
}

// Close closes the client.
// It must be called to shut down all connections to the etcd server.
func (c Client) Close() error {
	return c.c.Close()
}

// Options are the options for the etcd client.
type Options struct {
	// Addresses of the etcd servers in the cluster, including port.
	// Optional ([]string{"localhost:2379"} by default).
	Endpoints []string
	// The timeout for operations.
	// Optional (200 * time.Millisecond by default).
	Timeout *time.Duration
	// Encoding format.
	// Optional (encoding.JSON by default).
	Codec encoding.Codec
}

// DefaultOptions is an Options object with default values.
// Endpoints: []string{"localhost:2379"}, Timeout: 200 * time.Millisecond, Codec: encoding.JSON
var DefaultOptions = Options{
	Endpoints: []string{"localhost:2379"},
	Timeout:   &defaultTimeout,
	Codec:     encoding.JSON,
}

// NewClient creates a new etcd client.
//
// You must call the Close() method on the client when you're done working with it.
func NewClient(options Options) (Client, error) {
	result := Client{}

	// Set default values
	if options.Endpoints == nil || len(options.Endpoints) == 0 {
		options.Endpoints = DefaultOptions.Endpoints
	}
	if options.Timeout == nil {
		options.Timeout = DefaultOptions.Timeout
	}
	if options.Codec == nil {
		options.Codec = DefaultOptions.Codec
	}

	config := clientv3.Config{
		Endpoints:   options.Endpoints,
		DialTimeout: 2 * time.Second,
		DialOptions: []grpc.DialOption{grpc.WithBlock()},
	}

	cli, err := clientv3.New(config)
	if err != nil {
		return result, err
	}

	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	statusRes, err := cli.Status(ctxWithTimeout, options.Endpoints[0])
	if err != nil {
		return result, err
	} else if statusRes == nil {
		return result, errors.New("the status response from etcd was nil")
	}

	result.c = cli
	result.timeOut = *options.Timeout
	result.codec = options.Codec

	return result, nil
}
