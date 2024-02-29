package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/philippgille/gokv/encoding"
	"github.com/philippgille/gokv/util"
)

var setOpt = options.Replace().SetUpsert(true)

// item is the document that's stored in the MongoDB collection.
// mongo (un-)marshalls it to/from BSON automatically, when reading from / writing to MongoDB.
// In a previous version where we used github.com/globalsign/mgo, we had to do this
// in order to support all types that a gokv value can have, including strings and other simple values,
// because mgo only supported marshalling of structs and maps.
// See https://github.com/globalsign/mgo/blob/113d3961e7311526535a1ef7042196563d442761/bson/bson.go#L538.
// Now with the official MongoDB driver there's `bson.MarshalValue()` for simple values,
// but we still need to use a struct to be able to use the "_id" field as key (see below).
// And bson.MarshalValue() requires marshallers to be registered, whereas with our
// codecs the user can just define custom marshal methods on their types.
type item struct {
	// There are advantages and disavantages regarding the use of a string as "_id" instead of MongoDB's default ObjectId.
	// We can't use the ObjectId because we only have the key that the gokv package user passes us as parameter.
	// We could use a document with "_id" as ObjectId, "k" as string and "v" as slice of bytes and then create an MongoDB Index for "k".
	// That would have the advantage that we could activate the constraint that the indexed values must be unique
	// (which is not the case with the "_id" field, which is rarely realized due to the use of ObjectId as "_id"
	// and ObjectId being generated on the server to guarantee uniqueness).
	// But it would have the disadvantage that when clustering the MongoDB and sharding the MongoDB collection that we use,
	// the DB admin would have to use *our* indexed value as shard key, because otherwise it could lead to duplicate entries
	// even if the unique constraint is set. And the admin might not be aware of this. Using "_id" as shard key seems to be pretty standard.
	// At least that (advantages + disadvantages) is my understanding from the documentation and comments on Stackoverflow.
	// Relevant links:
	// - https://github.com/mongodb/docs/blob/5f2d5e7dce7766a14b25b0d032970f065a866110/source/core/document.txt#L78
	// - https://github.com/mongodb/docs/blob/e1b05bac8616fdfac13bedd79516a5ac33d4afdf/source/reference/bson-types.txt#L41
	// - https://github.com/mongodb/docs/blob/85171fd9fcc1cf2a5dc6f297b2b026c86bfbfd9d/source/indexes.txt#L46
	// - https://github.com/mongodb/docs/blob/81d03d2463bc995a451759ce44087fe7ecd4db74/source/core/sharding-shard-key.txt#L91
	K string `bson:"_id"`
	V []byte // "v" will be used as field name
}

// Client is a gokv.Store implementation for MongoDB.
type Client struct {
	c     *mongo.Collection
	codec encoding.Codec

	// Client and cancel are required on call to `Close()`
	client *mongo.Client
	cancel context.CancelFunc
}

// Gets underlying store to allow user manipulate object directly.
func (c Client) GetStore() *mongo.Collection {
	return c.c
}

// Set stores the given value for the given key.
// Values are automatically marshalled to JSON or gob (depending on the configuration).
// The key must not be "" and the value must not be nil.
func (c Client) Set(k string, v any) error {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return err
	}

	// First turn the passed object into something that MongoDB can handle
	data, err := c.codec.Marshal(v)
	if err != nil {
		return err
	}

	item := item{
		// K needs to be specified, otherwise an update operation (on an existing document)
		// would lead to the "_id" being overwritten by "",
		// which 1) we don't want of course and 2) leads to an error anyway.
		K: k,
		V: data,
	}
	_, err = c.c.ReplaceOne(context.Background(), bson.D{{"_id", k}}, item, setOpt)
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

	item := new(item)
	err = c.c.FindOne(context.Background(), bson.D{{"_id", k}}).Decode(item)
	// If no value was found return false
	if err == mongo.ErrNoDocuments {
		return false, nil
	} else if err != nil {
		return false, err
	}
	data := item.V

	return true, c.codec.Unmarshal(data, v)
}

// Delete deletes the stored value for the given key.
// Deleting a non-existing key-value pair does NOT lead to an error.
// The key must not be "".
func (c Client) Delete(k string) error {
	if err := util.CheckKey(k); err != nil {
		return err
	}

	_, err := c.c.DeleteOne(context.Background(), bson.D{{"_id", k}})
	// No need to check for mongo.ErrNoDocuments, because DeleteOne() doesn't return
	// any error if no document was deleted. This differs from a previous version
	// where we used mgo.
	return err
}

// Close closes the client.
// It must be called to release any open resources.
func (c Client) Close() error {
	c.cancel()
	return c.client.Disconnect(context.Background())
}

// Options are the options for the MongoDB client.
type Options struct {
	// Seed servers for the initial connection to the MongoDB cluster.
	// Format: [mongodb://][user:pass@]host1[:port1][,host2[:port2],...][/database][?options].
	// E.g.: "mongodb://localhost" (the port defaults to 27017).
	// Optional ("mongodb://localhost" by default).
	// For a detailed documentation and more examples see https://github.com/mongodb/docs/blob/01fa14decadc116b09ecdeae049e6744f16bf97f/source/reference/connection-string.txt.
	ConnectionString string
	// The name of the database to use.
	// Optional ("gokv" by default).
	DatabaseName string
	// The name of the collection to use.
	// Optional ("item" by default).
	CollectionName string
	// Encoding format.
	// Optional (encoding.JSON by default).
	Codec encoding.Codec
}

// DefaultOptions is an Options object with default values.
// ConnectionString: "localhost", DatabaseName: "gokv", CollectionName: "item", Codec: encoding.JSON
var DefaultOptions = Options{
	ConnectionString: "mongodb://localhost",
	DatabaseName:     "gokv",
	CollectionName:   "item",
	Codec:            encoding.JSON,
}

// NewClient creates a new MongoDB client.
//
// You must call the Close() method on the client when you're done working with it.
func NewClient(opts Options) (Client, error) {
	result := Client{}

	// Set default values
	if opts.ConnectionString == "" {
		opts.ConnectionString = DefaultOptions.ConnectionString
	}
	if opts.DatabaseName == "" {
		opts.DatabaseName = DefaultOptions.DatabaseName
	}
	if opts.CollectionName == "" {
		opts.CollectionName = DefaultOptions.CollectionName
	}
	if opts.Codec == nil {
		opts.Codec = DefaultOptions.Codec
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(opts.ConnectionString))
	if err != nil {
		return result, err
	}

	// The above `Connect` doesn't block for server discovery. But like with other
	// gokv store implementations, we want to ensure that after client creation
	// it's ready to use. So we call `Ping` to block until the server is discovered.
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer pingCancel()
	err = client.Ping(pingCtx, readpref.Primary())
	if err != nil {
		return result, err
	}

	c := client.Database(opts.DatabaseName).Collection(opts.CollectionName)

	result.c = c
	result.codec = opts.Codec
	result.client = client
	result.cancel = cancel

	return result, nil
}
