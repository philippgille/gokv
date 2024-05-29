package couchbase

import (
	"errors"
	"fmt"
	"time"

	"github.com/couchbase/gocb/v2"

	"github.com/philippgille/gokv/encoding"
	"github.com/philippgille/gokv/util"
)

var defaultTimeout = 2 * time.Second

// Client is a gokv.Store implementation for Couchbase.
type Client struct {
	collection *gocb.Collection
	cluster    *gocb.Cluster
	timeOut    time.Duration
	codec      encoding.Codec
	expiry     time.Duration
}

// Set will create or update the content of the given key on couchbase.
func (c *Client) Set(k string, v any) error {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return err
	}

	if c.codec != nil {
		data, err := c.codec.Marshal(v)
		if err != nil {
			return fmt.Errorf("codec: unable to marshal data: %w", err)
		}

		v = data
	}

	c.collection.Upsert(k, v, &gocb.UpsertOptions{
		Expiry:  c.expiry,
		Timeout: c.timeOut,
	})

	return nil
}

// Get will return the content of the given key from couchbase.
func (c *Client) Get(k string, v any) (bool, error) {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return false, err
	}

	if c.codec != nil {
		var data []byte

		found, err := c.rawGet(k, &data)
		if err != nil || !found {
			return found, err
		}

		return true, c.codec.Unmarshal(data, v)
	}

	return c.rawGet(k, v)
}

func (c *Client) rawGet(k string, v any) (bool, error) {
	docOut, err := c.collection.Get(k, &gocb.GetOptions{
		Timeout: c.timeOut,
	})
	if errors.Is(err, gocb.ErrDocumentNotFound) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	err = docOut.Content(v)
	if err != nil {
		return true, err
	}

	return true, nil
}

// Delete will remove the given key from couchbase.
func (c *Client) Delete(k string) error {
	if err := util.CheckKey(k); err != nil {
		return err
	}

	_, err := c.collection.Remove(k, &gocb.RemoveOptions{
		Timeout: c.timeOut,
	})

	if errors.Is(err, gocb.ErrDocumentNotFound) {
		return nil
	}

	return err
}

// Close shuts down all buckets in this cluster and invalidates any references this cluster has.
func (c *Client) Close() error {
	return c.cluster.Close(nil)
}

// Options are the options for the Couchbase client.
type Options struct {
	// ConnectionString is the couchbase connection string.
	ConnectionString string
	// The timeout for operations.
	// Optional (2 * time.Second by default).
	Timeout *time.Duration

	// Authenticator specifies the authenticator to use with the cluster.
	Authenticator gocb.Authenticator

	// Username & Password specifies the cluster username and password to
	// authenticate with.  This is equivalent to passing PasswordAuthenticator
	// as the Authenticator parameter with the same values
	Username string
	Password string

	// BucketName the name of the bucket to perform kv operations.
	BucketName string

	// ScopeName the name of the scope. Will use default scope by default.
	ScopeName string

	// CollectionName the name of the collection. Will use default collection by default.
	CollectionName string

	// Expiry will set the TTL of a given key on Set operations, if present.
	Expiry time.Duration

	// Codec accepts a given encoding.Codec to use on set / get operations.
	// If no Codec is set or it is encoding.JSON, we will use nothing, since
	// the default of couchbase is to use a JSON Transcoder.
	// By using encoding.Gob we will set a raw binary transcoder to be sure about
	// the flags and content on couchbase.
	Codec encoding.Codec
}

// DefaultOptions is an Options object with default values.
// ConnectionString: "couchbase://localhost". Timeout: 2 Seconds.
var DefaultOptions = Options{
	ConnectionString: "couchbase://localhost",
	Timeout:          &defaultTimeout,
}

// NewClient creates a new Couchbase client.
//
// You must call the Close() method on the client when you're done working with it.
func NewClient(options Options) (*Client, error) {
	if options.ConnectionString == "" {
		options.ConnectionString = DefaultOptions.ConnectionString
	}

	if options.Timeout == nil {
		options.Timeout = DefaultOptions.Timeout
	}
	var (
		transcoder gocb.Transcoder
		codec      encoding.Codec
	)

	switch options.Codec {
	case nil:
	case encoding.JSON:
		transcoder = gocb.NewJSONTranscoder()
	default:
		transcoder = gocb.NewRawBinaryTranscoder()
		codec = options.Codec
	}

	cluster, err := gocb.Connect(options.ConnectionString, gocb.ClusterOptions{
		Authenticator: options.Authenticator,
		Username:      options.Username,
		Password:      options.Password,
		Transcoder:    transcoder,
		TimeoutsConfig: gocb.TimeoutsConfig{
			ConnectTimeout: *options.Timeout,
			KVTimeout:      *options.Timeout,
		},
	})
	if err != nil {
		return nil, err
	}

	var collection *gocb.Collection

	bucket := cluster.Bucket(options.BucketName)

	if options.CollectionName != "" && options.ScopeName != "" {
		scope := bucket.Scope(options.ScopeName)
		collection = scope.Collection(options.CollectionName)
	} else if options.CollectionName != "" {
		collection = bucket.Collection(options.CollectionName)
	} else {
		collection = bucket.DefaultCollection()
	}

	return &Client{
		collection: collection,
		cluster:    cluster,
		timeOut:    *options.Timeout,
		codec:      codec,
		expiry:     options.Expiry,
	}, nil
}
