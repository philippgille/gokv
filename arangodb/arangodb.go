package arangodb

import (
	"context"
	"crypto/tls"
	"errors"
	"strings"

	"github.com/philippgille/gokv/encoding"
	"github.com/philippgille/gokv/util"

	"github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
)

// item is the document that's stored in the ArangoDB collection.
type item struct {
	// https://www.arangodb.com/docs/stable/data-modeling-naming-conventions-document-keys.html
	// arangodb key has very specific rules for the key
	// in short, it's a-zA-Z0-9  with characters
	// _ - : . @ ( ) + , = ; $ ! * ' %
	// allowed as well. importantly, arangodb does not support non utf-8 string.
	// since gokv does not make any promises, we must reencode the key
	//
	H string `json:"_key" velocypack:"_key"`

	// this is the unhashed key. users may create an index on it if they need to in an external program
	K string `json:"k" velocypack:"k"`

	// arangodb by default will store a byte array as an array of int64, each byte taking up 8 bytes of spcae
	// this is 8x the storage space required.
	// hex and base64 require 200% and 133% space when converted to a utf8 string (note that arangodb does not support non-utf8 strings)
	// the best way i know of therefore to store a binary blob, is to store a length prefixed int64 array.
	// https://github.com/arangodb/arangodb/issues/107#issuecomment-1071177815
	// note that if you view the document in aardvark, the inability of javascript to process numbers above 2^53 will make the data uninteractable through aardvark
	V []uint64 `json:"v" velocypack:"v"`
}

// Client is a gokv.Store implementation for ArangoDB.
type Client struct {
	c     driver.Collection
	codec encoding.Codec
}

// Set stores the given value for the given key.
// Values are automatically marshalled to JSON or gob (depending on the configuration).
// The key must not be "" and the value must not be nil.
func (c Client) Set(k string, v interface{}) error {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return err
	}
	// First turn the passed object into something that ArangoDB can handle
	data, err := c.codec.Marshal(v)
	if err != nil {
		return err
	}
	item := item{
		K: k,
		H: hashString(k),
		V: bytesToArray(data),
	}
	_, err = c.c.CreateDocument(context.TODO(), item)
	if err != nil {
		// if error is conflict, replace
		if driver.IsConflict(err) {
			_, err = c.c.ReplaceDocument(context.TODO(), item.H, item)
			if err != nil {
				return err
			}
			return nil
		}
		// otherwise, pass the error, since its not related
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
func (c Client) Get(k string, v interface{}) (found bool, err error) {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return false, err
	}
	item := new(item)
	_, err = c.c.ReadDocument(context.TODO(), hashString(k), item)
	// If no value was found return false
	if driver.IsNoMoreDocuments(err) || driver.IsNotFound(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	data := item.V
	return true, c.codec.Unmarshal(arrayToBytes(data), v)
}

// Delete deletes the stored value for the given key.
// Deleting a non-existing key-value pair does NOT lead to an error.
// The key must not be "".
func (c Client) Delete(k string) error {
	if err := util.CheckKey(k); err != nil {
		return err
	}
	_, err := c.c.RemoveDocument(context.TODO(), hashString(k))
	if err != nil && !driver.IsNotFound(err) {
		return err
	}
	return nil
}

// Close closes the client.
// As arangodb connects through http, there is no need to close
// Persistent connections are made through http2.
func (c Client) Close() error {
	return nil
}

// Options are the options for the ArangoDB client.
type Options struct {
	// Seed servers for the initial connection to the ArangoDB server/cluster.
	// comma separated list, e.g
	// E.g.: "http://localhost:8529,http://localhost:8530".
	// Optional ("http://localhost:8529" by default).
	Endpoints string
	// Arangodb username
	// Optional("root" by default)
	Username string
	// Arangodb Password
	// Optional ("" by default)
	Password string
	// The name of the database to use.
	// Optional ("gokv" by default).
	DatabaseName string
	// The name of the collection to use.
	// Optional ("item" by default).
	CollectionName string
	// Encoding format.
	// Optional (encoding.JSON by default).
	Codec encoding.Codec
	// Disable ssl verification (insecure)
	// ssl verification is ENABLED by default. set this to TRUE to DISABLE verification
	InsecureSkipVerify bool
}

// DefaultOptions is an Options object with default values.
// ConnectionString: "localhost", DatabaseName: "gokv", CollectionName: "item", Codec: encoding.JSON
var DefaultOptions = Options{
	Endpoints:          "http://localhost:8529",
	DatabaseName:       "gokv",
	CollectionName:     "item",
	Codec:              encoding.JSON,
	InsecureSkipVerify: false,
}

// NewClient creates a new ArangoDB client.
//
// You must call the Close() method on the client when you're done working with it.
func NewClient(options Options) (Client, error) {
	result := Client{}
	ctx := context.TODO()

	// Set default values
	if options.Endpoints == "" {
		options.Endpoints = DefaultOptions.Endpoints
	}
	if options.DatabaseName == "" {
		options.DatabaseName = DefaultOptions.DatabaseName
	}
	if options.CollectionName == "" {
		options.CollectionName = DefaultOptions.CollectionName
	}
	if options.Codec == nil {
		options.Codec = DefaultOptions.Codec
	}
	config := http.ConnectionConfig{
		Endpoints: strings.Split(options.Endpoints, ","),
	}
	if options.InsecureSkipVerify {
		config.TLSConfig = &tls.Config{
			InsecureSkipVerify: options.InsecureSkipVerify,
		}
	}
	for _, v := range config.Endpoints {
		if !(strings.HasPrefix(v, "http") || strings.HasPrefix(v, "tcp")) {
			return result, errors.New("no reachable servers")
		}
	}
	hp, err := http.NewConnection(config)
	if err != nil {
		return result, err
	}
	c, err := driver.NewClient(driver.ClientConfig{
		Connection:     hp,
		Authentication: driver.BasicAuthentication(options.Username, options.Password),
	})
	if err != nil {
		return result, err
	}
	ok, err := c.DatabaseExists(ctx, options.DatabaseName)
	if err != nil {
		return result, err
	}
	var db driver.Database
	if ok {
		db, err = c.Database(ctx, options.DatabaseName)
	} else {
		db, err = c.CreateDatabase(ctx, options.DatabaseName, &driver.CreateDatabaseOptions{})
	}
	if err != nil {
		return result, err
	}
	if result.c, err = ensureCollection(ctx, db, options.CollectionName); err != nil {
		return result, err
	}
	result.codec = options.Codec
	return result, nil
}

func ensureCollection(ctx context.Context, db driver.Database, name string) (c driver.Collection, err error) {
	ok, err := db.CollectionExists(ctx, name)
	if err != nil {
		return
	}
	if ok {
		c, err = db.Collection(ctx, name)
	} else {
		c, err = db.CreateCollection(ctx, name, &driver.CreateCollectionOptions{})
	}
	if err != nil {
		return
	}
	return c, nil
}
