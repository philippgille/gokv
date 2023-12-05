package bigcache

import (
	"context"
	"time"

	"github.com/allegro/bigcache/v3"

	"github.com/philippgille/gokv/encoding"
	"github.com/philippgille/gokv/util"
)

// Store is a gokv.Store implementation for BigCache.
type Store struct {
	s     *bigcache.BigCache
	codec encoding.Codec
}

// Set stores the given value for the given key.
// Values are automatically marshalled to JSON or gob (depending on the configuration).
// The key must not be "" and the value must not be nil.
func (s Store) Set(k string, v any) error {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return err
	}

	data, err := s.codec.Marshal(v)
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
func (s Store) Get(k string, v any) (found bool, err error) {
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

	return true, s.codec.Unmarshal(data, v)
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
	// Encoding format.
	// Optional (encoding.JSON by default).
	Codec encoding.Codec
}

// DefaultOptions is an Options object with default values.
// HardMaxCacheSize: 0 (no limit), Eviction: 0 (no limit), Codec: encoding.JSON
var DefaultOptions = Options{
	Codec: encoding.JSON,
	// No need to set Eviction or HardMaxCacheSize because their zero values are fine.
}

// NewStore creates a BigCache store.
//
// You should call the Close() method on the store when you're done working with it.
func NewStore(options Options) (Store, error) {
	result := Store{}

	// Set default options
	if options.Codec == nil {
		options.Codec = DefaultOptions.Codec
	}

	config := bigcache.DefaultConfig(options.Eviction)
	config.HardMaxCacheSize = options.HardMaxCacheSize
	cache, err := bigcache.New(context.Background(), config)
	if err != nil {
		return result, err
	}

	result.s = cache
	result.codec = options.Codec

	return result, nil
}
