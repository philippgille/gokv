package pudge

import (
	"errors"
	"github.com/philippgille/gokv/encoding"
	"github.com/philippgille/gokv/util"
	"github.com/recoilme/pudge"
	"time"
)

type Store struct {
	db    *pudge.Db
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

	return s.db.Set(k, data)
}

func (s Store) Get(k string, v interface{}) (found bool, err error) {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return false, err
	}

	var data []byte
	err = s.db.Get(k, &data)
	if errors.Is(err, pudge.ErrKeyNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, s.codec.Unmarshal(data, v)
}

func (s Store) Delete(k string) error {
	if err := util.CheckKey(k); err != nil {
		return err
	}

	err := s.db.Delete(k)
	if errors.Is(err, pudge.ErrKeyNotFound) {
		return nil
	} else {
		return err
	}
}

func (s Store) Close() error {
	return s.db.Close()
}

type StoreMode int

const (
	FileFirst   StoreMode = 0
	MemoryFirst StoreMode = 2
)

type Options struct {
	// Path of the DB file.
	// Optional ("pudge.db" by default).
	Path string
	// FSync interval, 0 - disable sync (os will sync, typically 30 sec or so).
	// Optional (30 seconds by default, 1 second at least).
	SyncInterval time.Duration
	// 0 - file first, 2 - memory first(with persist on close), 2 - with empty file - memory without persist
	StoreMode StoreMode
	// Creating file mode
	// Optional (0666 by default).
	FileMode int
	// Creating directories mode
	// Optional (0777 by default).
	DirMode int
	// Encoding format.
	// Optional (encoding.JSON by default).
	Codec encoding.Codec
}

// DefaultOptions is an Options object with default values.
var DefaultOptions = Options{
	Path:         "pudge.db",
	SyncInterval: 30 * time.Second,
	StoreMode:    FileFirst,
	FileMode:     0666,
	DirMode:      0777,
	Codec:        encoding.JSON,
}

// NewStore creates a new bbolt store.
func NewStore(options Options) (Store, error) {
	result := Store{}

	// Set default values
	if options.Path == "" {
		options.Path = DefaultOptions.Path
	}
	if options.FileMode == 0 {
		options.FileMode = 0666
	}
	if options.DirMode == 0 {
		options.DirMode = 0777
	}
	if options.Codec == nil {
		options.Codec = DefaultOptions.Codec
	}

	// Open DB
	db, err := pudge.Open(options.Path, &pudge.Config{
		FileMode:     options.FileMode,
		DirMode:      options.DirMode,
		SyncInterval: int(options.SyncInterval.Seconds()),
		StoreMode:    int(options.StoreMode),
	})
	if err != nil {
		return result, err
	}

	result.db = db
	result.codec = options.Codec

	return result, nil
}
