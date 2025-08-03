package freecache

import (
	"context"
	"github.com/coocood/freecache"

	"github.com/philippgille/gokv/encoding"
	"github.com/philippgille/gokv/util"
)

const minSize = 512 * 1024

// Store is a gokv.Store implementation for FreeCache.
type Store struct {
	s     *freecache.Cache
	codec encoding.Codec
}

// Set stores the given value for the given key.
// Values are automatically marshalled to JSON or gob (depending on the configuration).
// The key must not be "" and the value must not be nil.
func (s Store) Set(k string, v any) error {
	ctx := context.Background()
	return s.SetWithContext(ctx, k, v)
}

// SetWithContext is exactly like Set function just with added context as first argument.
func (s Store) SetWithContext(_ context.Context, k string, v any) error {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return err
	}

	data, err := s.codec.Marshal(v)
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
func (s Store) Get(k string, v any) (found bool, err error) {
	ctx := context.Background()
	return s.GetWithContext(ctx, k, v)
}

// GetWithContext is exactly like Get function just with added context as first argument.
func (s Store) GetWithContext(_ context.Context, k string, v any) (found bool, err error) {
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

	return true, s.codec.Unmarshal(data, v)
}

// Delete deletes the stored value for the given key.
// Deleting a non-existing key-value pair does NOT lead to an error.
// The key must not be "".
func (s Store) Delete(k string) error {
	ctx := context.Background()
	return s.DeleteWithContext(ctx, k)
}

// DeleteWithContext is exactly like Delete function just with added context as first argument.
func (s Store) DeleteWithContext(_ context.Context, k string) error {
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
	// TODO: Set s.s to nil to free up resources? "Resources" meaning the for example 256 MiB memory?
	return nil
}

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
	// Encoding format.
	// Optional (encoding.JSON by default).
	Codec encoding.Codec
}

// DefaultOptions is an Options object with default values.
// Size: 256 MiB, Codec: encoding.JSON
var DefaultOptions = Options{
	Size:  256 * 1024 * 1024,
	Codec: encoding.JSON,
}

// NewStore creates a FreeCache store.
//
// You should call the Close() method on the store when you're done working with it.
func NewStore(options Options) Store {
	// Set default values
	if options.Size == 0 {
		options.Size = DefaultOptions.Size
	} else if options.Size < minSize {
		options.Size = minSize
	}
	if options.Codec == nil {
		options.Codec = DefaultOptions.Codec
	}

	cache := freecache.NewCache(options.Size)

	return Store{
		s:     cache,
		codec: options.Codec,
	}
}
