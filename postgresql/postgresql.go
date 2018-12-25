package postgresql

import (
	gosql "database/sql"

	// Usually a blank import is enough as it calls the package's init() function and loads the driver,
	// but we'll use the package's ParseDNS() function so we make this an actual import.
	_ "github.com/lib/pq"

	"github.com/philippgille/gokv/encoding"
	"github.com/philippgille/gokv/sql"
)

const defaultDBname = "gokv"

// Client is a gokv.Store implementation for PostgreSQL.
type Client struct {
	*sql.Client
}

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
	// this is NOT the "same host" (except when running with `--network host`)).
	ConnectionURL string
	// Name of the table in which the key-value pairs are stored.
	// Optional ("Item" by default).
	TableName string
	// Limits the number of open connections to the PostgreSQL server.
	// -1 for no limit. 0 will lead to the default value (100) being set.
	// Optional (100 by default).
	MaxOpenConnections int
	// Encoding format.
	// Optional (encoding.JSON by default).
	Codec encoding.Codec
}

// DefaultOptions is an Options object with default values.
// ConnectionURL: "postgres://postgres@/gokv?sslmode=disable", TableName: "Item", MaxOpenConnections: 100, Codec: encoding.JSON
var DefaultOptions = Options{
	ConnectionURL:      "postgres://postgres@/" + defaultDBname + "?sslmode=disable",
	TableName:          "Item",
	MaxOpenConnections: 100,
	Codec:              encoding.JSON,
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
	if options.Codec == nil {
		options.Codec = DefaultOptions.Codec
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
