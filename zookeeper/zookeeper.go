package zookeeper

import (
	"errors"
	"strings"
	"time"

	"github.com/samuel/go-zookeeper/zk"

	"github.com/philippgille/gokv/encoding"
	"github.com/philippgille/gokv/util"
)

// Client is a gokv.Store implementation for Apache ZooKeeper.
type Client struct {
	c          *zk.Conn
	pathPrefix string
	codec      encoding.Codec
}

// Set stores the given value for the given key.
// Values are automatically marshalled to JSON or gob (depending on the configuration).
// The key must not be "" and the value must not be nil.
func (c Client) Set(k string, v any) error {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return err
	}

	// First turn the passed object into something that Apache ZooKeeper can handle
	data, err := c.codec.Marshal(v)
	if err != nil {
		return err
	}

	k = c.pathPrefix + k
	acl := zk.WorldACL(zk.PermAll)
	_, err = c.c.Create(k, data, 0, acl)
	if err != nil {
		if err.Error() == "zk: node already exists" {
			_, err = c.c.Set(k, data, -1)
		}
	}
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

	k = c.pathPrefix + k
	data, _, err := c.c.Get(k)
	if err != nil {
		if err.Error() == "zk: node does not exist" {
			return false, nil
		}
		return false, err
	}

	return true, c.codec.Unmarshal(data, v)
}

// Delete deletes the stored value for the given key.
// Deleting a non-existing key-value pair does NOT lead to an error.
// The key must not be "".
func (c Client) Delete(k string) error {
	if err := util.CheckKey(k); err != nil {
		return err
	}

	k = c.pathPrefix + k
	err := c.c.Delete(k, -1)
	if err != nil && err.Error() == "zk: node does not exist" {
		return nil
	}
	return err
}

// Close closes the client.
// It must be called to close the underlying ZooKeeper client.
func (c Client) Close() error {
	c.c.Close()
	return nil
}

// Options are the options for the Apache ZooKeeper client.
type Options struct {
	// Server addresses including their port.
	// Optional ("localhost:2181" by default).
	Servers []string
	// Path prefix to use for each key.
	// Must start with "/".
	// Can be used to store each value in a specific "directory".
	// Begin and end with "/" to use as "directory".
	// Optional ("/gokv/" by default).
	PathPrefix string
	// Encoding format.
	// Optional (encoding.JSON by default).
	Codec encoding.Codec
}

// DefaultOptions is an Options object with default values.
// Servers: "localhost:2181", PathPrefix: "/gokv/", Codec: encoding.JSON
var DefaultOptions = Options{
	Servers:    []string{"localhost:2181"},
	PathPrefix: "/gokv/",
	Codec:      encoding.JSON,
}

// NewClient creates a new Apache ZooKeeper client.
//
// You must call the Close() method on the client when you're done working with it.
func NewClient(options Options) (Client, error) {
	result := Client{}

	// Precondition check
	if options.PathPrefix != "" && !strings.HasPrefix(options.PathPrefix, "/") {
		return result, errors.New("the PathPrefix must start with a \\")
	}

	// Set default values
	if options.Servers == nil {
		options.Servers = DefaultOptions.Servers
	}
	if options.PathPrefix == "" {
		options.PathPrefix = DefaultOptions.PathPrefix
	}
	if options.Codec == nil {
		options.Codec = DefaultOptions.Codec
	}

	c, _, err := zk.Connect(options.Servers, 2*time.Second, zk.WithLogInfo(false))
	if err != nil {
		return result, err
	}

	// Check connection
	_, _, err = c.Children("/")
	if err != nil {
		return result, err
	}

	// Create node if it doesn't exist (or even multiple nodes).
	// Doesn't need to be done if the PathPrefix is "/".
	if options.PathPrefix != "/" {
		// Split examples:
		// "/foo" splits to ["", "foo"]
		// "/foo/" splits to ["", "foo", ""]
		// "/foo/bar" splits to ["", "foo", "bar"]
		baseNodes := strings.Split(options.PathPrefix, "/")
		// Don't care about the first element
		baseNodes = baseNodes[1:]
		// No need to create any nodes if length == 1 (at this point, after removing the first one),
		// because that's e.g. "/foo", which is no extra node, just a key prefix.
		if len(baseNodes) > 1 {
			// Also remove the last element, because there are two cases and we don't care about any of them:
			// 1) If it's "" it's not a new node
			// 2) If it's not "" it's just a prefix for a key
			baseNodes = baseNodes[:len(baseNodes)-1]
			nodeToCreate := "/"
			acl := zk.WorldACL(zk.PermAll)
			for _, pathElem := range baseNodes {
				// No path elem should be empty, because that would mean a PathPrefix containing "//" was used
				if pathElem == "" {
					return result, errors.New("invalid PathPrefix containing \"//\"")
				}
				nodeToCreate += pathElem
				_, _, err = c.Get(nodeToCreate)
				if err != nil {
					if err.Error() == "zk: node does not exist" {
						_, err = c.Create(nodeToCreate, nil, 0, acl)
						if err != nil {
							return result, err
						}
					} else {
						return result, err
					}
				}
				nodeToCreate += "/"
			}
		}
	}

	result.c = c
	result.pathPrefix = options.PathPrefix
	result.codec = options.Codec

	return result, nil
}
