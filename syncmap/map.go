package syncmap

import (
	"errors"
	"sync"

	"github.com/philippgille/gokv/util"
)

// Store is a gokv.Store implementation for a Go sync.Map.
type Store struct {
	m             *sync.Map
	marshalFormat MarshalFormat
}

// Set stores the given value for the given key.
// Values are automatically marshalled to JSON or gob (depending on the configuration).
// The key must not be "" and the value must not be nil.
func (m Store) Set(k string, v interface{}) error {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return err
	}

	var data []byte
	var err error
	switch m.marshalFormat {
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

	m.m.Store(k, data)
	return nil
}

// Get retrieves the stored value for the given key.
// You need to pass a pointer to the value, so in case of a struct
// the automatic unmarshalling can populate the fields of the object
// that v points to with the values of the retrieved object's values.
// If no value is found it returns (false, nil).
// The key must not be "" and the pointer must not be nil.
func (m Store) Get(k string, v interface{}) (bool, error) {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return false, err
	}

	data, found := m.m.Load(k)
	if !found {
		return false, nil
	}

	switch m.marshalFormat {
	case JSON:
		return true, util.FromJSON(data.([]byte), v)
	case Gob:
		return true, util.FromGob(data.([]byte), v)
	default:
		return true, errors.New("The store seems to be configured with a marshal format that's not implemented yet")
	}
}

// Delete deletes the stored value for the given key.
// Deleting a non-existing key-value pair does NOT lead to an error.
// The key must not be "".
func (m Store) Delete(k string) error {
	if err := util.CheckKey(k); err != nil {
		return err
	}

	m.m.Delete(k)
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

// Options are the options for the Go sync.Map store.
type Options struct {
	// (Un-)marshal format.
	// Optional (JSON by default).
	MarshalFormat MarshalFormat
}

// DefaultOptions is an Options object with default values.
// MarshalFormat: JSON
var DefaultOptions = Options{
	// No need to set MarshalFormat to JSON
	// because its zero value is fine.
}

// NewStore creates a new Go sync.Map store.
func NewStore(options Options) Store {
	return Store{
		m:             &sync.Map{},
		marshalFormat: options.MarshalFormat,
	}
}
