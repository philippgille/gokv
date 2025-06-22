package mysql

import (
	gosql "database/sql"

	// Usually a blank import is enough as it calls the package's init() function and loads the driver,
	// but we'll use the package's ParseDNS() function so we make this an actual import.
	gosqldriver "github.com/go-sql-driver/mysql"

	"github.com/philippgille/gokv/encoding"
	"github.com/philippgille/gokv/sql"
)

const (
	defaultDBname = "gokv"
	keyLength     = "255"
)

// It's a code smell to work with a hard coded number,
// but the error doesn't seem to be defined as constant or variable
// in neither of the two packages (database/sql and github.com/go-sql-driver/mysql).
const errDBnotFound = 1049

// Client is a gokv.Store implementation for MySQL.
type Client struct {
	c *sql.Client
}

// Set stores the given value for the given key.
// Values are automatically marshalled to JSON or gob (depending on the configuration).
// The length of the key must not exceed 255 characters.
// The key must not be "" and the value must not be nil.
func (c Client) Set(k string, v any) error {
	// It's tempting to remove this "wrapper" method
	// and just use *sql.Client as embedded field,
	// But we need this explicit method for a different GoDoc
	// (255 character limit).
	return c.c.Set(k, v)
}

// Get retrieves the stored value for the given key.
// You need to pass a pointer to the value, so in case of a struct
// the automatic unmarshalling can populate the fields of the object
// that v points to with the values of the retrieved object's values.
// If no value is found it returns (false, nil).
// The length of the key must not exceed 255 characters.
// The key must not be "" and the pointer must not be nil.
func (c Client) Get(k string, v any) (found bool, err error) {
	// It's tempting to remove this "wrapper" method
	// and just use *sql.Client as embedded field,
	// But we need this explicit method for a different GoDoc
	// (255 character limit).
	return c.c.Get(k, v)
}

// Delete deletes the stored value for the given key.
// Deleting a non-existing key-value pair does NOT lead to an error.
// The length of the key must not exceed 255 characters.
// The key must not be "".
func (c Client) Delete(k string) error {
	// It's tempting to remove this "wrapper" method
	// and just use *sql.Client as embedded field,
	// But we need this explicit method for a different GoDoc
	// (255 character limit).
	return c.c.Delete(k)
}

// Close closes the client.
// It must be called to return all open connections to the connection pool and to release any open resources.
func (c Client) Close() error {
	return c.c.Close()
}

// Options are the options for the MySQL client.
type Options struct {
	// Connection string.
	// Format: [username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN].
	// Full example: "username:password@protocol(address)/dbname?param=value".
	// Minimal example: "/dbname".
	// If you leave the dbname out of the connection string,
	// an attempt will be made to create a database named "gokv".
	// If you passed a name of a database that doesn't exist yet,
	// an attempt will be made to create it.
	// If the user doesn't have the permission to create databases, an error is returned.
	// Optional ("root@/gokv" by default, which will connect to "127.0.0.1:3306"
	// and requires the server to be configured with MYSQL_ALLOW_EMPTY_PASSWORD=true,
	// which should only be done in local test environments, if at all).
	DataSourceName string
	// Name of the table in which the key-value pairs are stored.
	// Optional ("Item" by default).
	TableName string
	// Limits the number of open connections to the MySQL server.
	// -1 for no limit. 0 will lead to the default value (100) being set.
	// Optional (100 by default).
	MaxOpenConnections int
	// Encoding format.
	// Optional (encoding.JSON by default).
	Codec encoding.Codec
}

// DefaultOptions is an Options object with default values.
// DataSourceName: "root@/gokv", TableName: "Item", MaxOpenConnections: 100, Codec: encoding.JSON
var DefaultOptions = Options{
	DataSourceName:     "root@/" + defaultDBname,
	TableName:          "Item",
	MaxOpenConnections: 100,
	Codec:              encoding.JSON,
}

// NewClient creates a new MySQL client.
//
// You must call the Close() method on the client when you're done working with it.
func NewClient(options Options) (Client, error) {
	result := Client{}

	// Set default values
	if options.DataSourceName == "" {
		options.DataSourceName = DefaultOptions.DataSourceName
	}
	if options.TableName == "" {
		options.TableName = DefaultOptions.TableName
	}
	switch options.MaxOpenConnections {
	case 0:
		options.MaxOpenConnections = DefaultOptions.MaxOpenConnections
	case -1:
		options.MaxOpenConnections = 0 // 0 actually leads to the MySQL driver using no connection limit.
	}
	if options.Codec == nil {
		options.Codec = DefaultOptions.Codec
	}

	db, err := gosql.Open("mysql", options.DataSourceName)
	if err != nil {
		return result, err
	}

	cfg, err := gosqldriver.ParseDSN(options.DataSourceName)
	if err != nil {
		return result, err
	}

	err = db.Ping()
	if err != nil {
		if driverErr, ok := err.(*gosqldriver.MySQLError); ok {
			// If the package user included a database name in the DataSourceName,
			// but the database doesn't exist yet, we can try to create + use that database.
			if driverErr.Number == errDBnotFound {
				err = createDB(cfg, db)
				if err != nil {
					return result, nil
				}
			} else {
				return result, err
			}
		} else {
			return result, err
		}
	} else if cfg.DBName == "" {
		// Ping() was successful, but in case the package user didn't include a database name
		// in the DataSourceName, we must now attempt to create the database.
		newDB, err := createDefaultDB(db, cfg)
		if err != nil {
			return result, err
		}
		// Also, we must replace the current value that the db pointer points to by the new db,
		// because calling "USE" on a database only works for the single connection that's used
		// instead of for all connections in the pool.
		db.Close()
		db = newDB
	}

	// Limit number of concurrent connections. Typical max connections on a MySQL server is 100.
	// This prevents "Error 1040: Too many connections", which otherwise occurs for example with 500 concurrent goroutines.
	db.SetMaxOpenConns(options.MaxOpenConnections)

	// Create table if it doesn't exist yet.
	//
	// TEXT can't be used as primary key. VARCHAR default length is 255.
	// VARCHAR with more characters are converted for example to SMALLTEXT,
	// which again can't be used as primary key.
	// TODO: Maybe the 255 character limit only applies to MySQL < 5.x?
	// If yes, allow the user to define a key length via the options.
	// Also: There's no hard character limit, but byte limit.
	// So the 255 characters come from 255 utf8mb3 characters.
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS " + options.TableName + " (k VARCHAR(" + keyLength + ") PRIMARY KEY, v BLOB NOT NULL)")
	if err != nil {
		return result, err
	}

	// Create prepared statements that will be reused for every Set()/Get() operation.
	// Note: Prepared statements are handled differently from other programming languages in Go,
	// see: http://go-database-sql.org/prepared.html.
	// TODO: Prepared statements might prevent the use of other databases that are compatible with the MySQL protocol.
	upsertStmt, err := db.Prepare("INSERT INTO " + options.TableName + " (k, v) VALUES (?, ?) ON DUPLICATE KEY UPDATE v = VALUES(v)")
	if err != nil {
		return result, err
	}
	getStmt, err := db.Prepare("SELECT v FROM " + options.TableName + " WHERE k = ?")
	if err != nil {
		return result, err
	}
	deleteStmt, err := db.Prepare("DELETE FROM " + options.TableName + " where k = ?")
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

	result.c = &c

	return result, nil
}

// createDB creates a DB when the DB name is given in the config.
func createDB(cfg *gosqldriver.Config, db *gosql.DB) error {
	// We can't use the existing db object, because that would lead to the same error again.
	// So create a new one without database name, but "backup" the user provided database name,
	// which we need to create the database.
	userProvidedDBname := cfg.DBName
	cfg.DBName = ""
	dsnWithoutDBname := cfg.FormatDSN()
	tempDB, err := gosql.Open("mysql", dsnWithoutDBname)
	// This temporary DB must be closed.
	// But let's not return an error in case closing this temporary DB fails.
	// TODO: Maybe DO return an error? If yes, also change GolangCI-Lint configuration to not exclude this warning.
	defer tempDB.Close()
	if err != nil {
		return err
	}
	err = tempDB.Ping()
	if err != nil {
		return err
	}
	// No need to check if userProvidedDBname == "", because in that case the error wouldn't be 1049 (unknown database).
	// In case the user doesn't have the permission to create a database, an error is returned.
	err = sql.CreateDB(tempDB, userProvidedDBname)
	if err != nil {
		return err
	}
	// Now the initial ping should work.
	err = db.Ping()
	if err != nil {
		return err
	}

	return nil
}

// createDefaultDB creates a DB with the default name
// and returns a pointer to the new sql.DB object.
// It's a new object and you should probably call Close() on the passed db object.
func createDefaultDB(db *gosql.DB, cfg *gosqldriver.Config) (*gosql.DB, error) {
	err := sql.CreateDB(db, defaultDBname)
	if err != nil {
		return nil, err
	}
	cfg.DBName = defaultDBname
	dsnWithDBname := cfg.FormatDSN()
	newDB, err := gosql.Open("mysql", dsnWithDBname)
	if err != nil {
		return nil, err
	}
	err = newDB.Ping()
	if err != nil {
		return nil, err
	}

	return newDB, nil
}
