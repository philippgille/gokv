package freecache

import (
	"errors"

	"github.com/coocood/freecache"

	"github.com/philippgille/gokv/util"
)

const minSize = 512 * 1024

// Store is a gokv.Store implementation for FreeCache.
type Store struct {
	s             *freecache.Cache
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

	return s.s.Set([]byte(k), data, 0)
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

	data, err := s.s.Get([]byte(k))
	if err != nil {
		if err == freecache.ErrNotFound {
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

	s.s.Del([]byte(k))
	return nil
}

// Close closes the store.
// When called, the cache is cleared.
func (s Store) Close() error {
	s.s.Clear()
	return nil
}

// MarshalFormat is an enum for the available (un-)marshal formats of this gokv.Store implementation.
type MarshalFormat int

const (
	// JSON is the MarshalFormat for (un-)marshalling to/from JSON
	JSON MarshalFormat = iota
	// Gob is the MarshalFormat for (un-)marshalling to/from gob
	Gob
)

// Options are the options for the FreeCache store.
type Options struct {
	// The size of the cache in bytes.
	// 512 KiB is the minimum size
	// (if you set a lower size, 512 KiB will be used instead).
	// If you set 0, the default size will be used.
	// When the size is reached and you store new entries,
	// old entries are evicted.
	// Optional (256 MiB by default).
	Size int
	// (Un-)marshal format.
	// Optional (JSON by default).
	MarshalFormat MarshalFormat
}

// DefaultOptions is an Options object with default values.
// Size: 256 MiB, MarshalFormat: JSON
var DefaultOptions = Options{
	Size: 256 * 1024 * 1024,
	// No need to set MarshalFormat to JSON
	// because its zero value is fine.
}

// NewStore creates a FreeCache store.
func NewStore(options Options) Store {
	// Set default values
	if options.Size == 0 {
		options.Size = DefaultOptions.Size
	} else if options.Size < minSize {
		options.Size = minSize
	}

	cache := freecache.NewCache(options.Size)

	return Store{
		s:             cache,
		marshalFormat: options.MarshalFormat,
	}
}
