package buntdb

import (
	"errors"
	"github.com/philippgille/gokv/encoding"
	"github.com/philippgille/gokv/util"
	"github.com/tidwall/buntdb"
)

type Store struct {
	db    *buntdb.DB
	codec encoding.Codec
}

func (s Store) Set(k string, v interface{}) error {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return err
	}

	data, err := s.codec.Marshal(v)
	if err != nil {
		return err
	}

	return s.db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(k, string(data), nil)
		return err
	})
}

func (s Store) Get(k string, v interface{}) (found bool, err error) {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return false, err
	}

	var val string
	err = s.db.View(func(tx *buntdb.Tx) error {
		val, err = tx.Get(k)
		return err
	})
	if errors.Is(err, buntdb.ErrNotFound) {
		return false, nil
	}

	return true, s.codec.Unmarshal([]byte(val), v)
}

func (s Store) Delete(k string) error {
	if err := util.CheckKey(k); err != nil {
		return err
	}

	err := s.db.Update(func(tx *buntdb.Tx) error {
		_, err := tx.Delete(k)
		return err
	})
	if errors.Is(err, buntdb.ErrNotFound) {
		return nil
	}

	return err
}

func (s Store) Close() error {
	return s.db.Close()
}

// Options are the options for the BuntDB instance.
type Options struct {
	// Path to database file.
	// Optional (":memory:" by default).
	Path string
	// Encoding format.
	// Optional (encoding.JSON by default).
	Codec encoding.Codec
}

var DefaultOptions = Options{
	Path:  ":memory:",
	Codec: encoding.JSON,
}

func NewStore(options Options) (Store, error) {
	result := Store{}

	// Set default values
	if options.Path == "" {
		options.Path = DefaultOptions.Path
	}
	if options.Codec == nil {
		options.Codec = DefaultOptions.Codec
	}

	db, err := buntdb.Open(options.Path)
	if err != nil {
		return result, err
	}

	result.db = db
	result.codec = options.Codec
	return result, nil
}
