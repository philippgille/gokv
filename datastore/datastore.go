package datastore

import (
	"context"
	"errors"
	"time"

	"cloud.google.com/go/datastore"
	"google.golang.org/api/option"

	"github.com/philippgille/gokv/encoding"
	"github.com/philippgille/gokv/util"
)

const kind = "gokv"

// entity is a struct that holds the actual value as a slice of bytes named "V"
// (translated to lowercase "v" in Cloud Datastore).
// Cloud Datastore requires a pointer to a struct as value.
// The key doesn't need to be part of the struct.
type entity struct {
	V []byte `datastore:"v,noindex"`
}

var defaultTimeout = 2 * time.Second

// Client is a gokv.Store implementation for Cloud Datastore.
type Client struct {
	c       *datastore.Client
	timeOut time.Duration
	codec   encoding.Codec
}

// Set stores the given value for the given key.
// Values are automatically marshalled to JSON or gob (depending on the configuration).
// The key must not be "" and the value must not be nil.
func (c Client) Set(k string, v any) error {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return err
	}

	// First turn the passed object into something that Cloud Datastore can handle.
	data, err := c.codec.Marshal(v)
	if err != nil {
		return err
	}

	tctx, cancel := context.WithTimeout(context.Background(), c.timeOut)
	defer cancel()
	key := datastore.Key{
		Kind: kind,
		Name: k,
	}
	src := entity{
		V: data,
	}
	_, err = c.c.Put(tctx, &key, &src)

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

	tctx, cancel := context.WithTimeout(context.Background(), c.timeOut)
	defer cancel()
	key := datastore.Key{
		Kind: kind,
		Name: k,
	}
	dst := new(entity)
	err = c.c.Get(tctx, &key, dst)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			return false, nil
		}
		return false, err
	}
	data := dst.V

	return true, c.codec.Unmarshal(data, v)
}

// Delete deletes the stored value for the given key.
// Deleting a non-existing key-value pair does NOT lead to an error.
// The key must not be "".
func (c Client) Delete(k string) error {
	if err := util.CheckKey(k); err != nil {
		return err
	}

	tctx, cancel := context.WithTimeout(context.Background(), c.timeOut)
	defer cancel()
	key := datastore.Key{
		Kind: kind,
		Name: k,
	}
	return c.c.Delete(tctx, &key)
}

// Close closes the client.
func (c Client) Close() error {
	return c.c.Close()
}

// Options are the options for the Cloud Datastore client.
type Options struct {
	// ID of the Google Cloud project.
	ProjectID string
	// Path to the credentials file. For example:
	// "/home/user/Downloads/[FILE_NAME].json".
	// If you don't set a credentials file explicitly,
	// the GCP SDK will look for the file path in the
	// GOOGLE_APPLICATION_CREDENTIALS environment variable.
	// Optional ("" by default, leading to a lookup via environment variable).
	CredentialsFile string
	// The timeout for operations.
	// Optional (2 * time.Second by default).
	Timeout *time.Duration
	// Encoding format.
	// Optional (encoding.JSON by default).
	Codec encoding.Codec
}

// DefaultOptions is an Options object with default values.
// CredentialsFile: "", Timeout: 2 * time.Second, Codec: encoding.JSON
var DefaultOptions = Options{
	Timeout: &defaultTimeout,
	Codec:   encoding.JSON,
	// No need to set CredentialsFile because its Go zero value is fine.
}

// NewClient creates a new Cloud Datastore client.
//
// You must call the Close() method on the store when you're done working with it.
func NewClient(options Options) (Client, error) {
	result := Client{}

	// Precondition check
	if options.ProjectID == "" {
		return result, errors.New("The ProjectID in the options must not be empty")
	}

	// Set default values
	if options.Timeout == nil {
		options.Timeout = DefaultOptions.Timeout
	}
	if options.Codec == nil {
		options.Codec = DefaultOptions.Codec
	}

	// Don't pass a context with timeout to NewClient,
	// because the Dial() call in NewClient() is non-blocking anyway
	// and it would interfere with credential refreshing.
	var dsClient *datastore.Client
	var err error
	if options.CredentialsFile == "" {
		dsClient, err = datastore.NewClient(context.Background(), options.ProjectID)
	} else {
		dsClient, err = datastore.NewClient(context.Background(), options.ProjectID, option.WithCredentialsFile(options.CredentialsFile))
	}
	if err != nil {
		return result, err
	}

	result.c = dsClient
	result.timeOut = *options.Timeout
	result.codec = options.Codec

	return result, nil
}
