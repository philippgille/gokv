package cockroachdb

import (
	gosql "database/sql"

	_ "github.com/lib/pq"

	"github.com/philippgille/gokv/encoding"
	"github.com/philippgille/gokv/sql"
)

const defaultDBname = "gokv"

// Client is a gokv.Store implementation for CockroachDB.
type Client struct {
	*sql.Client
}

// Options are the options for the CockroachDB client.
type Options struct {
	// Connection URL.
	// Format: postgres://username[:password]@address/dbname[?param1=value1&...&paramN=valueN].
	// Example: "postgres://roach:secret@localhost:26257/gokv?sslmode=disable".
	// The database ("dbname" in the example) must already exist.
	// For a full list of available connection paramters, see:
	// https://github.com/cockroachdb/docs/blob/560c4227f4d811c5be9dc8e4a5385e508d0c68e5/v2.1/connection-parameters.md#additional-connection-parameters.
	// Optional ("postgres://root@localhost:26257/gokv?sslmode=disable&application_name=gokv" by default,
	// which will connect to "localhost:26257" as root user and doesn't use TLS,
	// which you should NOT do in production).
	ConnectionURL string
	// Name of the table in which the key-value pairs are stored.
	// Optional ("Item" by default).
	TableName string
	// Limits the number of open connections to the CockroachDB server.
	// -1 for no limit. 0 will lead to the default value (100) being set.
	// Optional (100 by default).
	MaxOpenConnections int
	// Encoding format.
	// Optional (encoding.JSON by default).
	Codec encoding.Codec
}

// DefaultOptions is an Options object with default values.
// ConnectionURL: "postgres://root@localhost:26257/gokv?sslmode=disable&application_name=gokv", TableName: "Item", MaxOpenConnections: 100, Codec: encoding.JSON
var DefaultOptions = Options{
	ConnectionURL:      "postgres://root@localhost:26257/" + defaultDBname + "?sslmode=disable&application_name=gokv",
	TableName:          "Item",
	MaxOpenConnections: 100,
	Codec:              encoding.JSON,
}

// NewClient creates a new CockroachDB client.
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

	// Limit number of concurrent connections.
	// Typical max connections on a PostgreSQL server is 100,
	// not sure what the typical value is for CockroachDB.
	// This prevents "Error 1040: Too many connections"
	// (or similar, I think this specific error message is from MySQL),
	// which otherwise occurs for example with 500 concurrent goroutines.
	db.SetMaxOpenConns(options.MaxOpenConnections)

	// Create table if it doesn't exist yet.
	// Use a "column family" so that a the row is a single entry in CockroachDB's underlying key-value store.
	// See: https://forum.cockroachlabs.com/t/can-i-use-cockroachdb-as-a-kv-store/56.
	// And: https://github.com/cockroachdb/docs/blob/b68c9ad8097d1efec4d2b6d849f6788a0e857215/v2.1/column-families.md
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS " + options.TableName + " (k STRING PRIMARY KEY, v BYTES NOT NULL, FAMILY kv (k, v))")
	if err != nil {
		return result, err
	}

	// Create prepared statements that will be reused for every Set()/Get() operation.
	// Note: Prepared statements are handled differently from other programming languages in Go,
	// see: http://go-database-sql.org/prepared.html.

	// Use "UPSERT" instead of "INSERT ON CONFLICT" because it doesn't do a read operation to determine the write operation,
	// which makes it faster.
	upsertStmt, err := db.Prepare("UPSERT INTO " + options.TableName + " (k, v) VALUES ($1, $2)")
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
