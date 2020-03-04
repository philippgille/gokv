package consul

import (
	"github.com/hashicorp/consul/api"
	"sync"

	"github.com/philippgille/gokv/encoding"
	"github.com/philippgille/gokv/util"
)

// Client is a gokv.Store implementation for Consul.
type Client struct {
	c      *api.KV
	folder string
	codec  encoding.Codec
	m      *sync.Mutex
}

// Set stores the given value for the given key.
// Values are automatically marshalled to JSON or gob (depending on the configuration).
// The key must not be "" and the value must not be nil.
func (c Client) Set(k string, v interface{}) error {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return err
	}

	// First turn the passed object into something that Consul can handle
	data, err := c.codec.Marshal(v)
	if err != nil {
		return err
	}

	if c.folder != "" {
		k = c.folder + "/" + k
	}

	return c.set(k, data)
}

func (c Client) set(k string, data []byte) error {
	c.m.Lock()
	defer c.m.Unlock()

	kvPair := api.KVPair{
		Key:   k,
		Value: data,
	}

	_, err := c.c.Put(&kvPair, nil)
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

	if c.folder != "" {
		k = c.folder + "/" + k
	}

	kvPair, err := c.get(k)
	if err != nil {
		return false, err
	}

	// If no value was found return false
	if kvPair == nil {
		return false, nil
	}
	data := kvPair.Value

	return true, c.codec.Unmarshal(data, v)
}

func (c Client) get(k string) (*api.KVPair, error) {
	c.m.Lock()
	defer c.m.Unlock()

	kvPair, _, err := c.c.Get(k, nil)
	return kvPair, err
}

// Delete deletes the stored value for the given key.
// Deleting a non-existing key-value pair does NOT lead to an error.
// The key must not be "".
func (c Client) Delete(k string) error {
	if err := util.CheckKey(k); err != nil {
		return err
	}

	if c.folder != "" {
		k = c.folder + "/" + k
	}

	return c.delete(k)
}

func (c Client) delete(k string) error {
	c.m.Lock()
	defer c.m.Unlock()

	_, err := c.c.Delete(k, nil)
	return err
}

// Close closes the client.
// In the Consul implementation this doesn't have any effect.
func (c Client) Close() error {
	return nil
}

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
	// Encoding format.
	// Optional (encoding.JSON by default).
	Codec encoding.Codec
}

// DefaultOptions is an Options object with default values.
// Scheme: "http", Address: "127.0.0.1:8500", Folder: none, Codec: encoding.JSON
var DefaultOptions = Options{
	Scheme:  "http",
	Address: "127.0.0.1:8500",
	Codec:   encoding.JSON,
	// No need to define Folder because its zero value is fine
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
	if options.Codec == nil {
		options.Codec = DefaultOptions.Codec
	}

	config := api.DefaultConfig()
	config.Scheme = options.Scheme
	config.Address = options.Address
	client, err := api.NewClient(config)
	if err != nil {
		return result, err
	}

	result.c = client.KV()
	result.folder = options.Folder
	result.codec = options.Codec
	result.m = &sync.Mutex{}

	return result, nil
}
