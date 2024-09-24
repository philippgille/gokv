package sqlite

import (
	gosql "database/sql"

	_ "modernc.org/sqlite"

	"github.com/philippgille/gokv/encoding"
	"github.com/philippgille/gokv/sql"
)

const defaultDBname = "gokv"

// Client is a gokv.Store implementation for sqlite.
type Client struct {
	*sql.Client
}

// Options are the options for the sqlite client.
type Options struct {
	// Path
	Path string
	// Name of the table in which the key-value pairs are stored.
	// Optional ("Item" by default).
	TableName string
	// Encoding format.
	// Optional (encoding.JSON by default).
	Codec encoding.Codec
}

// DefaultOptions is an Options object with default values.
var DefaultOptions = Options{
	Path:      "sqlite.db",
	TableName: defaultDBname,
	Codec:     encoding.JSON,
}

// NewClient creates a new sqlite client.
//
// You must call the Close() method on the client when you're done working with it.
func NewClient(options Options) (Client, error) {
	result := Client{}

	// Set default values
	if options.TableName == "" {
		options.TableName = DefaultOptions.TableName
	}
	if options.Codec == nil {
		options.Codec = DefaultOptions.Codec
	}
	if options.Path == "" {
		options.Path = DefaultOptions.Path
	}

	db, err := gosql.Open("sqlite", options.Path)
	if err != nil {
		return result, err
	}

	err = db.Ping()
	if err != nil {
		return result, err
	}

	const q = `
	PRAGMA foreign_keys = ON;
	PRAGMA synchronous = NORMAL;
	PRAGMA journal_mode = 'WAL';
	PRAGMA cache_size = -64000;
	`

	_, err = db.Exec(q)

	db.SetMaxOpenConns(1)

	if err != nil {
		return result, err
	}

	// Create table if it doesn't exist yet.
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS " + options.TableName + " (k TEXT PRIMARY KEY, v BLOB NOT NULL)")
	if err != nil {
		return result, err
	}

	// Create prepared statements that will be reused for every Set()/Get() operation.
	// Note: Prepared statements are handled differently from other programming languages in Go,
	// see: http://go-database-sql.org/prepared.html.
	upsertStmt, err := db.Prepare("INSERT INTO " + options.TableName + " (k, v) VALUES ($1, $2) ON CONFLICT (k) DO UPDATE SET v = $2")
	if err != nil {
		return result, err
	}
	getStmt, err := db.Prepare("SELECT v FROM " + options.TableName + " WHERE k = $1")
	if err != nil {
		return result, err
	}
	deleteStmt, err := db.Prepare("DELETE FROM " + options.TableName + " where k = $1")
	if err != nil {
		return result, err
	}

	c := sql.Client{
		C:          db,
		UpsertStmt: upsertStmt,
		GetStmt:    getStmt,
		DeleteStmt: deleteStmt,
		Codec:      options.Codec,
	}

	result.Client = &c

	return result, nil
}

// Close closes the client.
// It must be called to make sure that all open transactions finish and to release all DB resources.
func (c Client) Close() error {
	return c.Client.Close()
}

// Get retrieves the stored value for the given key.
func (c Client) Get(k string, v any) (found bool, err error) {
	return c.Client.Get(k, v)
}

// Delete deletes the stored value for the given key.
func (c Client) Delete(k string) error {
	return c.Client.Delete(k)
}

func (c Client) Set(k string, v any) error {
	return c.Client.Set(k, v)
}
