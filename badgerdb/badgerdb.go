package badgerdb

import (
	"github.com/dgraph-io/badger"

	"github.com/philippgille/gokv/util"
)

// Store is a gokv.Store implementation for BadgerDB.
type Store struct {
	db *badger.DB
}

// Set stores the given value for the given key.
// Values are marshalled to JSON automatically.
func (c Store) Set(k string, v interface{}) error {
	// First turn the passed object into something that BadgerDB can handle
	data, err := util.ToJSON(v)
	if err != nil {
		return err
	}

	err = c.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(k), data)
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
	var data []byte
	err := c.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(k))
		if err != nil {
			return err
		}
		// item.Value() is only valid within the transaction.
		// We can either copy it ourselves or use the ValueCopy() method.
		data, err = item.ValueCopy(nil)
		if err != nil {
			return err
		}
		return nil
	})
	// If no value was found return false
	if err == badger.ErrKeyNotFound {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, util.FromJSON(data, v)
}

// Delete deletes the stored value for the given key.
// Deleting a non-existing key-value pair does NOT lead to an error.
func (c Store) Delete(k string) error {
	return c.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(k))
	})
}

// Options are the options for the BadgerDB store.
type Options struct {
	// Directory for storing the DB files.
	// Optional ("BadgerDB" by default).
	Dir string
}

// DefaultOptions is an Options object with default values.
// Dir: "BadgerDB"
var DefaultOptions = Options{
	Dir: "BadgerDB",
}

// NewStore creates a new BadgerDB store.
// Note: BadgerDB uses an exclusive write lock on the database directory so it cannot be shared by multiple processes.
// So when creating multiple clients you should always use a new database directory (by setting a different Path in the options).
//
// Don't worry about closing the BadgerDB as long as you don't need to close the DB while the process that opened it runs.
func NewStore(options Options) (Store, error) {
	result := Store{}

	// Set default values
	if options.Dir == "" {
		options.Dir = DefaultOptions.Dir
	}

	// Open the Badger database located in the options.Dir directory.
	// It will be created if it doesn't exist.
	opts := badger.DefaultOptions
	opts.Dir = options.Dir
	opts.ValueDir = opts.Dir
	db, err := badger.Open(opts)
	if err != nil {
		return result, err
	}

	result = Store{
		db: db,
	}

	return result, nil
}
