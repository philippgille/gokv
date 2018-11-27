package etcd

import (
	"context"
	"errors"
	"time"

	"go.etcd.io/etcd/clientv3"

	"github.com/philippgille/gokv/util"
)

// Client is a gokv.Store implementation for etcd.
type Client struct {
	c             *clientv3.Client
	marshalFormat MarshalFormat
}

// Set stores the given value for the given key.
// Values are automatically marshalled to JSON or gob (depending on the configuration).
// The key must not be "" and the value must not be nil.
func (c Client) Set(k string, v interface{}) error {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return err
	}

	// First turn the passed object into something that etcd can handle
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

	_, err = c.c.Put(context.Background(), k, string(data))
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
func (c Client) Get(k string, v interface{}) (bool, error) {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return false, err
	}

	getRes, err := c.c.Get(context.Background(), k)
	if err != nil {
		return false, err
	}
	kvs := getRes.Kvs
	// If no value was found return false
	if kvs == nil || len(kvs) == 0 {
		return false, nil
	}
	data := kvs[0].Value

	switch c.marshalFormat {
	case JSON:
		return true, util.FromJSON(data, v)
	case Gob:
		return true, util.FromGob(data, v)
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

	_, err := c.c.Delete(context.Background(), k)
	return err
}

// Close closes the client.
// It must be called to shut down all connections to the etcd server.
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

// Options are the options for the etcd client.
type Options struct {
	// Addresses of the etcd servers in the cluster, including port.
	// Optional ([]string{"localhost:2379"} by default).
	Endpoints []string
	// (Un-)marshal format.
	// Optional (JSON by default).
	MarshalFormat MarshalFormat
}

// DefaultOptions is an Options object with default values.
// Endpoints: []string{"localhost:2379"}, MarshalFormat: JSON
var DefaultOptions = Options{
	Endpoints: []string{"localhost:2379"},
	// No need to set MarshalFormat to JSON because its zero value is fine.
}

// NewClient creates a new etcd client.
func NewClient(options Options) (Client, error) {
	result := Client{}

	// Set default values
	if options.Endpoints == nil || len(options.Endpoints) == 0 {
		options.Endpoints = DefaultOptions.Endpoints
	}

	// The behaviour for New() seems to be inconsistent.
	// It should block at most for the specified time in DialTimeout.
	// In our case though New() doesn't block, but instead the following call does.
	// Maybe it's just the specific version we're using.
	// See https://github.com/etcd-io/etcd/issues/9829#issuecomment-438434795.
	// Use own timeout as workaround.
	// TODO: Remove workaround after etcd behaviour has been fixed or clarified.
	//cli, err := clientv3.NewFromURLs(options.Endpoints)
	config := clientv3.Config{
		Endpoints:   options.Endpoints,
		DialTimeout: 2 * time.Second,
	}
	errChan := make(chan error, 1)
	cliChan := make(chan *clientv3.Client, 1)
	go func() {
		cli, err := clientv3.New(config)
		if err != nil {
			errChan <- err
			return
		}
		statusRes, err := cli.Status(context.Background(), options.Endpoints[0])
		if err != nil {
			errChan <- err
			return
		}
		if statusRes == nil {
			errChan <- errors.New("The status response from etcd was nil")
			return
		}
		cliChan <- cli
	}()
	select {
	case err := <-errChan:
		return result, err
	case cli := <-cliChan:
		result = Client{
			c:             cli,
			marshalFormat: options.MarshalFormat,
		}
		return result, nil
	case <-time.After(3 * time.Second):
		return result, errors.New("A timeout occurred while trying to connect to the etcd server")
	}
}
