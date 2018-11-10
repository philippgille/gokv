package consul

import (
	"errors"

	"github.com/hashicorp/consul/api"

	"github.com/philippgille/gokv/util"
)

// Client is a gokv.Store implementation for Consul.
type Client struct {
	c             *api.KV
	folder        string
	marshalFormat MarshalFormat
}

// Set stores the given value for the given key.
// Values are marshalled to JSON automatically.
func (c Client) Set(k string, v interface{}) error {
	// First turn the passed object into something that Consul can handle
	var data []byte
	var err error
	switch c.marshalFormat {
	case JSON:
		data, err = util.ToJSON(v)
	case Gob:
		data, err = util.ToGob(v)
	default:
		return errors.New("The store seems to be configured with a marshal format that's not implemented yet")
	}
	if err != nil {
		return err
	}

	if c.folder != "" {
		k = c.folder + "/" + k
	}
	kvPair := api.KVPair{
		Key:   k,
		Value: data,
	}
	_, err = c.c.Put(&kvPair, nil)
	if err != nil {
		return err
	}

	return nil
}

// Get retrieves the stored value for the given key.
// You need to pass a pointer to the value, so in case of a struct
// the automatic unmarshalling can populate the fields of the object
// that v points to with the values of the retrieved object's values.
func (c Client) Get(k string, v interface{}) (bool, error) {
	if c.folder != "" {
		k = c.folder + "/" + k
	}
	kvPair, _, err := c.c.Get(k, nil)
	if err != nil {
		return false, err
	}
	// If no value was found return false
	if kvPair == nil {
		return false, nil
	}
	data := kvPair.Value

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
func (c Client) Delete(k string) error {
	if c.folder != "" {
		k = c.folder + "/" + k
	}
	_, err := c.c.Delete(k, nil)
	return err
}

// MarshalFormat is an enum for the available (un-)marshal formats of this gokv.Store implementation.
type MarshalFormat int

const (
	// JSON is the MarshalFormat for (un-)marshalling to/from JSON
	JSON MarshalFormat = iota
	// Gob is the MarshalFormat for (un-)marshalling to/from gob
	Gob
)

// Options are the options for the Consul client.
type Options struct {
	// URI scheme for the Consul server.
	// Optional ("http" by default).
	Scheme string
	// Address of the Consul server, including port number.
	// Optional ("127.0.0.1:8500" by default).
	Address string
	// Directory under which to store the key-value pairs.
	// The Consul UI calls this "folder".
	// Optional (none by default).
	Folder string
	// (Un-)marshal format.
	// Optional (JSON by default).
	MarshalFormat MarshalFormat
}

// DefaultOptions is an Options object with default values.
// Scheme: "http", Address: "127.0.0.1:8500", Folder: none, MarshalFormat: JSON
var DefaultOptions = Options{
	Scheme:  "http",
	Address: "127.0.0.1:8500",
	// No need to define Folder because its zero value is fine
	// No need to set MarshalFormat to JSON because its zero value is fine.
}

// NewClient creates a new Consul client.
func NewClient(options Options) (Client, error) {
	result := Client{}

	// Set default values
	if options.Scheme == "" {
		options.Scheme = DefaultOptions.Scheme
	}
	if options.Address == "" {
		options.Address = DefaultOptions.Address
	}

	config := api.DefaultConfig()
	config.Scheme = options.Scheme
	config.Address = options.Address
	client, err := api.NewClient(config)
	if err != nil {
		return result, err
	}

	result = Client{
		c:             client.KV(),
		folder:        options.Folder,
		marshalFormat: options.MarshalFormat,
	}

	return result, nil
}
