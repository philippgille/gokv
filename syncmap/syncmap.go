package syncmap

import (
	"context"
	"sync"

	"github.com/philippgille/gokv/encoding"
	"github.com/philippgille/gokv/util"
)

// Store is a gokv.Store implementation for a Go sync.Map.
type Store struct {
	m     *sync.Map
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

	s.m.Store(k, data)
	return nil
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

	dataInterface, found := s.m.Load(k)
	if !found {
		return false, nil
	}
	// No need to check "ok" return value in type assertion,
	// because we control the map and we only put slices of bytes in the map.
	data := dataInterface.([]byte)

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

	s.m.Delete(k)
	return nil
}

// Close closes the store.
// When called, the store's pointer to the internal Go map is set to nil,
// leading to the map being free for garbage collection.
func (s Store) Close() error {
	s.m = nil
	return nil
}

// Options are the options for the Go sync.Map store.
type Options struct {
	// Encoding format.
	// Optional (encoding.JSON by default).
	Codec encoding.Codec
}

// DefaultOptions is an Options object with default values.
// Codec: encoding.JSON
var DefaultOptions = Options{
	Codec: encoding.JSON,
}

// NewStore creates a new Go sync.Map store.
//
// You should call the Close() method on the store when you're done working with it.
func NewStore(options Options) Store {
	// Set default values
	if options.Codec == nil {
		options.Codec = DefaultOptions.Codec
	}

	return Store{
		m:     &sync.Map{},
		codec: options.Codec,
	}
}
