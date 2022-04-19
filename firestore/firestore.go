package firestore

import (
	"context"
	"errors"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/philippgille/gokv/encoding"
	"github.com/philippgille/gokv/util"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const collection = "gokv"

// Client is a gokv.Store implementation for Cloud Firestore.
type Client struct {
	firestore *firestore.Client
	codec     encoding.Codec
}

// Set stores the given value for the given key.
// Values are automatically marshalled to JSON or gob (depending on the configuration).
// The key must not be "" and the value must not be nil.
func (c Client) Set(k string, v interface{}) error {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return err
	}

	// First turn the passed object into something that Cloud Datastore can handle.
	data, err := c.codec.Marshal(v)
	if err != nil {
		return err
	}

	tctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err = c.firestore.Collection(collection).Doc(k).Set(tctx, map[string]interface{}{
		"value": data,
	})
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

	d, err := c.firestore.Collection(collection).Doc(k).Get(tctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return false, nil
		}
		return false, err
	}
	data := (d.Data()["value"]).([]byte)
	return true, c.codec.Unmarshal(data, v)
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

	_, err := c.firestore.Collection(collection).Doc(k).Delete(tctx)
	return err
}

// Close closes the client.
func (c Client) Close() error {
	return c.firestore.Close()
}

// Options are the options for the Cloud Firestore client.
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

	// Encoding format.
	// Optional (encoding.JSON by default).
	Codec encoding.Codec
}

// DefaultOptions is an Options object with default values.
// CredentialsFile: "", Codec: encoding.JSON
var DefaultOptions = Options{
	Codec: encoding.JSON,
	// No need to set CredentialsFile because its Go zero value is fine.
}

// NewClient creates a new Cloud Firestore client.
//
// You must call the Close() method on the store when you're done working with it.
func NewClient(options Options) (Client, error) {
	result := Client{}

	// Precondition check
	if options.ProjectID == "" {
		return result, errors.New("the ProjectID in the options must not be empty")
	}

	// Set default values
	if options.Codec == nil {
		options.Codec = DefaultOptions.Codec
	}

	var firestoreClient *firestore.Client
	var err error

	if options.CredentialsFile == "" {
		firestoreClient, err = firestore.NewClient(context.Background(), options.ProjectID)
	} else {
		firestoreClient, err = firestore.NewClient(context.Background(), options.ProjectID, option.WithCredentialsFile(options.CredentialsFile))
	}
	if err != nil {
		return result, err
	}

	result.firestore = firestoreClient
	result.codec = options.Codec

	return result, nil
}
