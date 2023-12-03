package leveldb

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"

	"github.com/philippgille/gokv/encoding"
	"github.com/philippgille/gokv/util"
)

// Store is a gokv.Store implementation for LevelDB.
type Store struct {
	db        *leveldb.DB
	writeSync bool
	codec     encoding.Codec
}

// Set stores the given value for the given key.
// Values are automatically marshalled to JSON or gob (depending on the configuration).
// The key must not be "" and the value must not be nil.
func (s Store) Set(k string, v any) error {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return err
	}

	// First turn the passed object into something that LevelDB can handle
	data, err := s.codec.Marshal(v)
	if err != nil {
		return err
	}

	var writeOptions *opt.WriteOptions
	if s.writeSync {
		writeOptions = &opt.WriteOptions{
			Sync: true,
		}
	}
	return s.db.Put([]byte(k), data, writeOptions)
}

// Get retrieves the stored value for the given key.
// You need to pass a pointer to the value, so in case of a struct
// the automatic unmarshalling can populate the fields of the object
// that v points to with the values of the retrieved object's values.
// If no value is found it returns (false, nil).
// The key must not be "" and the pointer must not be nil.
func (s Store) Get(k string, v any) (found bool, err error) {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return false, err
	}

	data, err := s.db.Get([]byte(k), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return false, nil
		}
		return false, err
	}

	return true, s.codec.Unmarshal(data, v)
}

// Delete deletes the stored value for the given key.
// Deleting a non-existing key-value pair does NOT lead to an error.
// The key must not be "".
func (s Store) Delete(k string) error {
	if err := util.CheckKey(k); err != nil {
		return err
	}

	var writeOptions *opt.WriteOptions
	if s.writeSync {
		writeOptions = &opt.WriteOptions{
			Sync: true,
		}
	}
	return s.db.Delete([]byte(k), writeOptions)
}

// Close closes the store.
// It must be called to releases any outstanding snapshots,
// abort any in-flight compactions and discard open transactions.
func (s Store) Close() error {
	return s.db.Close()
}

// Options are the options for the LevelDB store.
type Options struct {
	// Path of the DB files.
	// Optional ("leveldb" by default).
	Path string
	// Flag to enable immediate file synchronization on writes.
	// If enabled, writes take longer, but no writes are lost when the system crashes.
	// If disabled, writes go to a cache first and are persisted via snapshots automatically.
	// Set() and Delete() are both writes.
	// Optional (false by default).
	WriteSync bool
	// Encoding format.
	// Optional (encoding.JSON by default).
	Codec encoding.Codec
}

// DefaultOptions is an Options object with default values.
// Path: "leveldb", WriteSync: false, Codec: encoding.JSON
var DefaultOptions = Options{
	Path:  "leveldb",
	Codec: encoding.JSON,
	// No need to set WriteSync because its zero values is fine.
}

// NewStore creates a new LevelDB store.
//
// You must call the Close() method on the store when you're done working with it.
func NewStore(options Options) (Store, error) {
	result := Store{}

	// Set default values
	if options.Path == "" {
		options.Path = DefaultOptions.Path
	}
	if options.Codec == nil {
		options.Codec = DefaultOptions.Codec
	}

	// Open DB
	db, err := leveldb.OpenFile(options.Path, nil)
	if err != nil {
		return result, err
	}

	result.db = db
	result.writeSync = options.WriteSync
	result.codec = options.Codec

	return result, nil
}
