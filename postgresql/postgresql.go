package postgresql

import (
	gosql "database/sql"

	// Usually a blank import is enough as it calls the package's init() function and loads the driver,
	// but we'll use the package's ParseDNS() function so we make this an actual import.
	_ "github.com/lib/pq"

	"github.com/philippgille/gokv/sql"
)

const defaultDBname = "gokv"

// Client is a gokv.Store implementation for PostgreSQL.
type Client struct {
	*sql.Client
}

// MarshalFormat is an enum for the available (un-)marshal formats of this gokv.Store implementation.
type MarshalFormat int

const (
	// JSON is the MarshalFormat for (un-)marshalling to/from JSON
	JSON MarshalFormat = iota
	// Gob is the MarshalFormat for (un-)marshalling to/from gob
	Gob
)

// Options are the options for the PostgreSQL client.
type Options struct {
	// Connection URL.
	// Format: postgres://username[:password]@[address]/dbname[?param1=value1&...&paramN=valueN].
	// Full example: "postgres://username:password@host:123/dbname?sslmode=verify-full".
	// Minimal example: "postgres://postgres@/dbname".
	// The database ("dbname" in the example) must already exist.
	// Optional ("postgres://postgres@/gokv?sslmode=disable" by default,
	// which will connect to "localhost:5432"
	// and requires the server to be configured with "trust" authentication
	// to not require a password when connecting from the same host.
	// When running the official PostgreSQL Docker container
	// and accessing it from outside the container,
	// this doesn't count as "same host").
	ConnectionURL string
	// Name of the table in which the key-value pairs are stored.
	// Optional ("Item" by default).
	TableName string
	// Limits the number of open connections to the PostgreSQL server.
	// -1 for no limit. 0 will lead to the default value (100) being set.
	// Optional (100 by default).
	MaxOpenConnections int
	// (Un-)marshal format.
	// Optional (JSON by default).
	MarshalFormat MarshalFormat
}

// DefaultOptions is an Options object with default values.
// ConnectionURL: "postgres://postgres@/gokv", TableName: "Item", MaxOpenConnections: 100, MarshalFormat: JSON
var DefaultOptions = Options{
	ConnectionURL:      "postgres://postgres@/" + defaultDBname,
	TableName:          "Item",
	MaxOpenConnections: 100,
	// No need to set MarshalFormat to JSON because its zero value is fine.
}

// NewClient creates a new PostgreSQL client.
//
// You must call the Close() method on the client when you're done working with it.
func NewClient(options Options) (Client, error) {
	result := Client{}

	// Set default values
	if options.ConnectionURL == "" {
		options.ConnectionURL = DefaultOptions.ConnectionURL
	}
	if options.TableName == "" {
		options.TableName = DefaultOptions.TableName
	}
	if options.MaxOpenConnections == 0 {
		options.MaxOpenConnections = DefaultOptions.MaxOpenConnections
	} else if options.MaxOpenConnections == -1 {
		options.MaxOpenConnections = 0 // 0 actually leads to the PostgreSQL driver using no connection limit.
	}

	db, err := gosql.Open("postgres", options.ConnectionURL)
	if err != nil {
		return result, err
	}

	err = db.Ping()
	if err != nil {
		return result, err
	}

	// Limit number of concurrent connections. Typical max connections on a PostgreSQL server is 100.
	// This prevents "Error 1040: Too many connections", which otherwise occurs for example with 500 concurrent goroutines.
	db.SetMaxOpenConns(options.MaxOpenConnections)

	// Create table if it doesn't exist yet.
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS " + options.TableName + " (k TEXT PRIMARY KEY, v BYTEA NOT NULL)")
	if err != nil {
		return result, err
	}

	// Create prepared statements that will be reused for every Set()/Get() operation.
	// Note: Prepared statements are handled differently from other programming languages in Go,
	// see: http://go-database-sql.org/prepared.html.
	// TODO: Prepared statements might prevent the use of other databases that are compatible with the PostgreSQL protocol.
	insertStmt, err := db.Prepare("INSERT INTO " + options.TableName + " (k, v) VALUES ($1, $2) ON CONFLICT (k) DO UPDATE SET v = $2")
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
		InsertStmt: insertStmt,
		GetStmt:    getStmt,
		DeleteStmt: deleteStmt,
		// TODO: This cast requires the order of the enum values to be the same. Fix with #47.
		MarshalFormat: sql.MarshalFormat(options.MarshalFormat),
	}

	result.Client = &c

	return result, nil
}