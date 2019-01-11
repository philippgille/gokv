package bigcache

import (
	"errors"
	"time"

	"github.com/allegro/bigcache"

	"github.com/philippgille/gokv/util"
)

const minSize = 512 * 1024

// Store is a gokv.Store implementation for BigCache.
type Store struct {
	s             *bigcache.BigCache
	marshalFormat MarshalFormat
}

// Set stores the given value for the given key.
// Values are automatically marshalled to JSON or gob (depending on the configuration).
// The key must not be "" and the value must not be nil.
func (s Store) Set(k string, v interface{}) error {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return err
	}

	var data []byte
	var err error
	switch s.marshalFormat {
	case JSON:
		data, err = util.ToJSON(v)
	case Gob:
		data, err = util.ToGob(v)
	default:
		err = errors.New("The store seems to be configured with a marshal format that's not implemented yet")
	}
	if err != nil {
		return err
	}

	return s.s.Set(k, data)
}

// Get retrieves the stored value for the given key.
// You need to pass a pointer to the value, so in case of a struct
// the automatic unmarshalling can populate the fields of the object
// that v points to with the values of the retrieved object's values.
// If no value is found it returns (false, nil).
// The key must not be "" and the pointer must not be nil.
func (s Store) Get(k string, v interface{}) (found bool, err error) {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return false, err
	}

	data, err := s.s.Get(k)
	if err != nil {
		if err == bigcache.ErrEntryNotFound {
			return false, nil
		}
		return false, err
	}

	switch s.marshalFormat {
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
// The key must not be "".
func (s Store) Delete(k string) error {
	if err := util.CheckKey(k); err != nil {
		return err
	}

	err := s.s.Delete(k)
	if err != nil {
		if err == bigcache.ErrEntryNotFound {
			return nil
		}
		return err
	}
	return nil
}

// Close closes the store.
// When called, the cache is left for removal by the garbage collector.
func (s Store) Close() error {
	return s.s.Close()
}

// MarshalFormat is an enum for the available (un-)marshal formats of this gokv.Store implementation.
type MarshalFormat int

const (
	// JSON is the MarshalFormat for (un-)marshalling to/from JSON
	JSON MarshalFormat = iota
	// Gob is the MarshalFormat for (un-)marshalling to/from gob
	Gob
)

// Options are the options for the BigCache store.
type Options struct {
	// The maximum size of the cache in MiB.
	// 0 means no limit.
	// Optional (0 by default, meaning no limit).
	HardMaxCacheSize int
	// Time after which an entry can be evicted.
	// 0 means no eviction.
	// When this is set to 0 and HardMaxCacheSize is set to a non-zero value
	// and the maximum capacity of the cache is reached
	// the oldest entries will be evicted nonetheless when new ones are stored.
	// Optional (0 by default, meaning no eviction).
	Eviction time.Duration
	// (Un-)marshal format.
	// Optional (JSON by default).
	MarshalFormat MarshalFormat
}

// DefaultOptions is an Options object with default values.
// HardMaxCacheSize: 0 (no limit), Eviction: 0 (no limit), MarshalFormat: JSON
var DefaultOptions = Options{
	// No need to set Eviction, HardMaxCacheSize or MarshalFormat
	// because their zero values are fine.
}

// NewStore creates a BigCache store.
//
// You should call the Close() method on the store when you're done working with it.
func NewStore(options Options) (Store, error) {
	result := Store{}

	config := bigcache.DefaultConfig(options.Eviction)
	config.HardMaxCacheSize = options.HardMaxCacheSize
	cache, err := bigcache.NewBigCache(config)
	if err != nil {
		return result, err
	}

	result.s = cache
	result.marshalFormat = options.MarshalFormat

	return result, nil
}
