package bolt

import (
	"errors"

	bolt "github.com/etcd-io/bbolt"

	"github.com/philippgille/gokv/util"
)

// Store is a gokv.Store implementation for bbolt (formerly known as Bolt / Bolt DB).
type Store struct {
	db            *bolt.DB
	bucketName    string
	marshalFormat MarshalFormat
}

// Set stores the given value for the given key.
// Values are marshalled to JSON automatically.
func (c Store) Set(k string, v interface{}) error {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return err
	}

	// First turn the passed object into something that Bolt can handle
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

	err = c.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(c.bucketName))
		return b.Put([]byte(k), data)
	})
	if err != nil {
		return err
	}
	return nil
}

// Get retrieves the stored value for the given key.
// You need to pass a pointer to the value, so in case of a struct
// the automatic unmarshalling can populate the fields of the object
// that v points to with the values of the retrieved object's values.
func (c Store) Get(k string, v interface{}) (bool, error) {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return false, err
	}

	var data []byte
	c.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(c.bucketName))
		txData := b.Get([]byte(k))
		// txData is only valid during the transaction.
		// Its value must be copied to make it valid outside of the tx.
		// TODO: Benchmark if it's faster to copy + close tx,
		// or to keep the tx open until unmarshalling is done.
		if txData != nil {
			// `data = append([]byte{}, txData...)` would also work, but the following is more explicit
			data = make([]byte, len(txData))
			copy(data, txData)
		}
		return nil
	})

	// If no value was found return false
	if data == nil {
		return false, nil
	}

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
func (c Store) Delete(k string) error {
	if err := util.CheckKey(k); err != nil {
		return err
	}

	return c.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(c.bucketName))
		return b.Delete([]byte(k))
	})
}

// MarshalFormat is an enum for the available (un-)marshal formats of this gokv.Store implementation.
type MarshalFormat int

const (
	// JSON is the MarshalFormat for (un-)marshalling to/from JSON
	JSON MarshalFormat = iota
	// Gob is the MarshalFormat for (un-)marshalling to/from gob
	Gob
)

// Options are the options for the Bolt store.
type Options struct {
	// Bucket name for storing the key-value pairs.
	// Optional ("default" by default).
	BucketName string
	// Path of the DB file.
	// Optional ("bolt.db" by default).
	Path string
	// (Un-)marshal format.
	// Optional (JSON by default).
	MarshalFormat MarshalFormat
}

// DefaultOptions is an Options object with default values.
// BucketName: "default", Path: "bolt.db", MarshalFormat: JSON
var DefaultOptions = Options{
	BucketName: "default",
	Path:       "bolt.db",
	// No need to set MarshalFormat to JSON
	// because its zero value is fine.
}

// NewStore creates a new Bolt store.
// Note: Bolt uses an exclusive write lock on the database file so it cannot be shared by multiple processes.
// So when creating multiple clients you should always use a new database file (by setting a different Path in the options).
//
// Don't worry about closing the Bolt DB as long as you don't need to close the DB while the process that opened it runs.
func NewStore(options Options) (Store, error) {
	result := Store{}

	// Set default values
	if options.BucketName == "" {
		options.BucketName = DefaultOptions.BucketName
	}
	if options.Path == "" {
		options.Path = DefaultOptions.Path
	}

	// Open DB
	db, err := bolt.Open(options.Path, 0600, nil)
	if err != nil {
		return result, err
	}

	// Create a bucket if it doesn't exist yet.
	// In Bolt key/value pairs are stored to and read from buckets.
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(options.BucketName))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return result, err
	}

	result = Store{
		db:            db,
		bucketName:    options.BucketName,
		marshalFormat: options.MarshalFormat,
	}

	return result, nil
}
