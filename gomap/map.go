package gomap

import (
	"sync"

	"github.com/philippgille/gokv/util"
)

// Store is a gokv.Store implementation for a Go map with a sync.RWMutex for concurrent access.
type Store struct {
	m    map[string][]byte
	lock *sync.RWMutex
}

// Set stores the given object for the given key.
// Values are marshalled to JSON automatically.
func (m Store) Set(k string, v interface{}) error {
	data, err := util.ToJSON(v)
	if err != nil {
		return err
	}
	m.lock.Lock()
	defer m.lock.Unlock()
	m.m[k] = data
	return nil
}

// Get retrieves the stored value for the given key.
// You need to pass a pointer to the value, so in case of a struct
// the automatic unmarshalling can populate the fields of the object
// that v points to with the values of the retrieved object's values.
func (m Store) Get(k string, v interface{}) (bool, error) {
	m.lock.RLock()
	data, found := m.m[k]
	// Unlock right after reading instead of with defer(),
	// because following unmarshalling will take some time
	// and we don't want to block writing threads until that's done.
	m.lock.RUnlock()
	if !found {
		return false, nil
	}

	return true, util.FromJSON(data, v)
}

// NewStore creates a new Go sync.Map store.
func NewStore() Store {
	return Store{
		m:    make(map[string][]byte),
		lock: new(sync.RWMutex),
	}
}
