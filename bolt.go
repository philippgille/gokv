package gokv

import (
	"sync"

	bolt "github.com/coreos/bbolt"
)

var bucketName = "ln-paywall"

// BoltClient is a StorageClient implementation for bbolt (formerly known as Bolt / Bolt DB).
type BoltClient struct {
	db   *bolt.DB
	lock *sync.Mutex
}

// Set stores the given object for the given key.
func (c BoltClient) Set(k string, v interface{}) error {
	// First turn the passed object into something that Bolt can handle
	data, err := toJSON(v)
	if err != nil {
		return err
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	err = c.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		err := b.Put([]byte(k), data)
		return err
	})
	if err != nil {
		return err
	}
	return nil
}

// Get retrieves the stored object for the given key and populates the fields of the object that v points to
// with the values of the retrieved object's values.
func (c BoltClient) Get(k string, v interface{}) (bool, error) {
	var data []byte
	err := c.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		data = b.Get([]byte(k))
		return nil
	})
	if err != nil {
		return false, err
	}

	// If no value was found assign nil to the pointer
	if data == nil {
		return false, nil
	}

	return true, fromJSON(data, v)
}

// BoltOptions are the options for the BoltClient.
type BoltOptions struct {
	// Path of the DB file.
	// Optional ("ln-paywall.db" by default).
	Path string
}

// DefaultBoltOptions is a BoltOptions object with default values.
// Path: "ln-paywall.db"
var DefaultBoltOptions = BoltOptions{
	Path: "ln-paywall.db",
}

// NewBoltClient creates a new BoltClient.
// Note: Bolt uses an exclusive write lock on the database file so it cannot be shared by multiple processes.
// For preventing clients from cheating (reusing preimages across different endpoints / middlewares that use
// different Bolt DB files) and for the previous mentioned reason you should use only one BoltClient.
// For example:
//  // ...
//  storageClient, err := storage.NewBoltClient(storage.DefaultBoltOptions) // Uses file "ln-paywall.db"
//  if err != nil {
//      panic(err)
//  }
//  cheapPaywall := wall.NewGinMiddleware(cheapInvoiceOptions, lnClient, storageClient)
//  expensivePaywall := wall.NewGinMiddleware(expensiveInvoiceOptions, lnClient, storageClient)
//  router.GET("/ping", cheapPaywall, pingHandler)
//  router.GET("/compute", expensivePaywall, computeHandler)
//  // ...
// If you want to start an additional web service, this would be an additional process, so you can't use the same
// DB file. You should look into the other storage options in this case, for example Redis.
//
// Don't worry about closing the Bolt DB, the middleware opens it once and uses it for the duration of its lifetime.
// When the web service is stopped, the DB file lock is released automatically.
func NewBoltClient(boltOptions BoltOptions) (BoltClient, error) {
	result := BoltClient{}

	// Set default values
	if boltOptions.Path == "" {
		boltOptions.Path = DefaultBoltOptions.Path
	}

	// Open DB
	db, err := bolt.Open(boltOptions.Path, 0600, nil)
	if err != nil {
		return result, err
	}

	// Create a bucket if it doesn't exist yet.
	// In Bolt key/value pairs are stored to and read from buckets.
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return result, err
	}

	result = BoltClient{
		db:   db,
		lock: &sync.Mutex{},
	}

	return result, nil
}
