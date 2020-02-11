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

type SyncPolicy = buntdb.SyncPolicy

const (
	Never       = buntdb.Never
	EverySecond = buntdb.EverySecond
	Always      = buntdb.Always
)

// Options are the options for the BuntDB instance.
type Options struct {
	// Path to database file.
	// Optional (":memory:" by default).
	Path string

	// SyncPolicy adjusts how often the data is synced to disk.
	// This value can be Never, EverySecond, or Always.
	// Optional (EverySecond by default).
	SyncPolicy SyncPolicy

	// AutoShrinkPercentage is used by the background process to trigger
	// a shrink of the aof file when the size of the file is larger than the
	// percentage of the result of the previous shrunk file.
	// For example, if this value is 100, and the last shrink process
	// resulted in a 100mb file, then the new aof file must be 200mb before
	// a shrink is triggered.
	// Optional (100% by default).
	AutoShrinkPercentage int

	// AutoShrinkMinSize defines the minimum size of the aof file before
	// an automatic shrink can occur.
	// Optional (32KB by default).
	AutoShrinkMinSize int

	// AutoShrinkDisabled turns off automatic background shrinking
	// Optional (false by default).
	AutoShrinkDisabled bool

	// Encoding format.
	// Optional (encoding.JSON by default).
	Codec encoding.Codec
}

var DefaultOptions = Options{
	Path:                 ":memory:",
	SyncPolicy:           EverySecond,
	AutoShrinkPercentage: 100,
	AutoShrinkMinSize:    32 * 1024 * 1024,
	AutoShrinkDisabled:   false,
	Codec:                encoding.JSON,
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

	err = db.SetConfig(buntdb.Config{
		SyncPolicy:           options.SyncPolicy,
		AutoShrinkPercentage: options.AutoShrinkPercentage,
		AutoShrinkMinSize:    options.AutoShrinkMinSize,
		AutoShrinkDisabled:   options.AutoShrinkDisabled,
	})
	if err != nil {
		return result, err
	}

	result.db = db
	result.codec = options.Codec
	return result, nil
}
