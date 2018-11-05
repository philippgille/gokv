package syncmap

import (
	"sync"

	"github.com/philippgille/gokv/util"
)

// Store is a gokv.Store implementation for a Go sync.Map.
type Store struct {
	m *sync.Map
}

// Set stores the given object for the given key.
// Values are marshalled to JSON automatically.
func (m Store) Set(k string, v interface{}) error {
	data, err := util.ToJSON(v)
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
func (m Store) Get(k string, v interface{}) (bool, error) {
	data, found := m.m.Load(k)
	if !found {
		return false, nil
	}

	return true, util.FromJSON(data.([]byte), v)
}

// Delete deletes the stored value for the given key.
func (m Store) Delete(k string) error {
	m.m.Delete(k)
	return nil
}

// NewStore creates a new Go sync.Map store.
func NewStore() Store {
	return Store{
		m: &sync.Map{},
	}
}
