package gokv

import (
	bolt "github.com/coreos/bbolt"
)

// BoltClient is a gokv.Store implementation for bbolt (formerly known as Bolt / Bolt DB).
type BoltClient struct {
	db         *bolt.DB
	bucketName string
}

// Set stores the given object for the given key.
func (c BoltClient) Set(k string, v interface{}) error {
	// First turn the passed object into something that Bolt can handle
	data, err := toJSON(v)
	if err != nil {
		return err
	}

	err = c.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(c.bucketName))
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
		b := tx.Bucket([]byte(c.bucketName))
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
	// Bucket name for storing the key-value pairs.
	// Optional ("default" by default).
	BucketName string
	// Path of the DB file.
	// Optional ("bolt.db" by default).
	Path string
}

// DefaultBoltOptions is a BoltOptions object with default values.
// BucketName: "default", Path: "bolt.db"
var DefaultBoltOptions = BoltOptions{
	BucketName: "default",
	Path:       "bolt.db",
}

// NewBoltClient creates a new BoltClient.
// Note: Bolt uses an exclusive write lock on the database file so it cannot be shared by multiple processes.
// So when creating multiple Bolt clients you should always use a new database file (by setting a different Path in the BoltOptions).
//
// Don't worry about closing the Bolt DB as long as you don't need to close the DB while the process that opened it runs.
func NewBoltClient(boltOptions BoltOptions) (BoltClient, error) {
	result := BoltClient{}

	// Set default values
	if boltOptions.BucketName == "" {
		boltOptions.BucketName = DefaultBoltOptions.BucketName
	}
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
		_, err := tx.CreateBucketIfNotExists([]byte(boltOptions.BucketName))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return result, err
	}

	result = BoltClient{
		db:         db,
		bucketName: boltOptions.BucketName,
	}

	return result, nil
}
