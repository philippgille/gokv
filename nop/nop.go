package nop

import "github.com/philippgille/gokv/util"

// Store is a gokv.Store implementation that does nothing except validate the arguments if applicable.
type Store struct{}

// Set pretends if stores the key. Always return nil error unless the key or value are invalid.
func (s Store) Set(k string, v interface{}) error {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return err
	}

	return nil
}

// Get pretends it fetches the key. Always return not found and nil error unless the key or value are invalid.
func (s Store) Get(k string, v interface{}) (found bool, err error) {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return false, err
	}

	return false, nil
}

// Delete pretends it deletes the key. Always return nil error unless the key is invalid.
func (s Store) Delete(k string) error {
	if err := util.CheckKey(k); err != nil {
		return err
	}

	return nil
}

// Close pretends it closes the store. Always return nil error.
func (s Store) Close() error {
	return nil
}

// NewStore creates a new nop Store that implements gokv.Store interface.
func NewStore() Store {
	return Store{}
}
