package datastore

import (
	"context"
	"errors"
	"time"

	"cloud.google.com/go/datastore"
	"google.golang.org/api/option"

	"github.com/philippgille/gokv/util"
)

const kind = "gokv"

// entity is a struct that holds the actual value as a slice of bytes named "V"
// (translated to lowercase "v" in Cloud Datastore).
// Cloud Datastore requires a pointer to a struct as value.
// The key doesn't need to be part of the struct.
type entity struct {
	V []byte `datastore: "v,noindex"`
}

// Client is a gokv.Store implementation for Cloud Datastore.
type Client struct {
	c             *datastore.Client
	marshalFormat MarshalFormat
}

// Set stores the given value for the given key.
// Values are automatically marshalled to JSON or gob (depending on the configuration).
// The key must not be "" and the value must not be nil.
func (c Client) Set(k string, v interface{}) error {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return err
	}

	// First turn the passed object into something that Cloud Datastore can handle.
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

	tctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
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
func (c Client) Get(k string, v interface{}) (found bool, err error) {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return false, err
	}

	tctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
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

	switch c.marshalFormat {
	case JSON:
		return true, util.FromJSON([]byte(data), v)
	case Gob:
		return true, util.FromGob([]byte(data), v)
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

	tctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
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

// MarshalFormat is an enum for the available (un-)marshal formats of this gokv.Store implementation.
type MarshalFormat int

const (
	// JSON is the MarshalFormat for (un-)marshalling to/from JSON
	JSON MarshalFormat = iota
	// Gob is the MarshalFormat for (un-)marshalling to/from gob
	Gob
)

// Options are the options for the Cloud Datastore client.
type Options struct {
	// ID of the Google Cloud project.
	ProjectID string
	// Path to the credentials file. For example:
	// "/home/user/Downloads/[FILE_NAME].json".
	// If you don't set a credentials file explicitly,
	// the GCP SDP will look for the file path in the
	// GOOGLE_APPLICATION_CREDENTIALS environment variable.
	// Optional ("" by default, leading to a lookup via environment variable).
	CredentialsFile string
	// (Un-)marshal format.
	// Optional (JSON by default).
	MarshalFormat MarshalFormat
}

// DefaultOptions is an Options object with default values.
// CredentialsFile: "", MarshalFormat: JSON
var DefaultOptions = Options{
	// No need to set CredentialsFile or MarshalFormat because their Go zero values are fine.
}

// NewClient creates a new Cloud Datastore client.
//
// You must call the Close() method on the store when you're done working with it.
func NewClient(options Options) (Client, error) {
	result := Client{}

	// Set default values
	if options.ProjectID == "" {
		return result, errors.New("The ProjectID in the options must not be empty")
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
	result.marshalFormat = options.MarshalFormat

	return result, nil
}
