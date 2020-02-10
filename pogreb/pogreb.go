package pogreb

import (
	"github.com/akrylysov/pogreb"
	"github.com/akrylysov/pogreb/fs"
	"github.com/philippgille/gokv/encoding"
	"github.com/philippgille/gokv/util"
	"time"
)

type Store struct {
	db    *pogreb.DB
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

	return s.db.Put([]byte(k), data)
}

func (s Store) Get(k string, v interface{}) (found bool, err error) {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return false, err
	}

	data, err := s.db.Get([]byte(k))
	if err != nil {
		return false, err
	}
	if data == nil {
		return false, nil
	}

	return true, s.codec.Unmarshal(data, v)
}

func (s Store) Delete(k string) error {
	if err := util.CheckKey(k); err != nil {
		return err
	}
	return s.db.Delete([]byte(k))
}

func (s Store) Close() error {
	return s.db.Close()
}

// Options are the options for the Pogreb store.
type Options struct {
	// Path of the DB file.
	// Optional ("pogreb.db" by default).
	Path string
	// FSync interval.
	// Optional (30 seconds by default).
	SyncInterval time.Duration
	// FileSystem represents a filesystem
	// Optional (fs.OS by default).
	FileSystem fs.FileSystem
	// Encoding format.
	// Optional (encoding.JSON by default).
	Codec encoding.Codec
}

// DefaultOptions is an Options object with default values.
var DefaultOptions = Options{
	Path:         "pogreb.db",
	SyncInterval: 30 * time.Second,
	FileSystem:   fs.OS,
	Codec:        encoding.JSON,
}

func NewStore(options Options) (Store, error) {
	result := Store{}

	// Set default values
	if options.Path == "" {
		options.Path = DefaultOptions.Path
	}
	if options.FileSystem == nil {
		options.FileSystem = DefaultOptions.FileSystem
	}
	if options.Codec == nil {
		options.Codec = DefaultOptions.Codec
	}

	err := util.CreateAllDirs(options.Path, 0777)
	if err != nil {
		return result, err
	}

	// Open DB
	db, err := pogreb.Open(options.Path, &pogreb.Options{
		BackgroundSyncInterval: options.SyncInterval,
		FileSystem:             options.FileSystem,
	})
	if err != nil {
		return result, err
	}

	result.db = db
	result.codec = options.Codec

	return result, nil
}
