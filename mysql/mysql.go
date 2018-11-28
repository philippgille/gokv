package mysql

import (
	"errors"

	"database/sql"

	// Usually a blank import is enough as it calls the package's init() function and loads the driver,
	// but we'll use the package's ParseDNS() function so we make this an actual import.
	gosqldriver "github.com/go-sql-driver/mysql"

	"github.com/philippgille/gokv/util"
)

const defaultDBname = "gokv"
const keyLength = "255"

// It's a code smell to work with a hard coded number,
// but the error doesn't seem to be defined as constant or variable
// in neither of the two packages (database/sql and github.com/go-sql-driver/mysql).
const errDBnotFound = 1049

// Client is a gokv.Store implementation for MySQL.
type Client struct {
	c             *sql.DB
	insertStmt    *sql.Stmt
	getStmt       *sql.Stmt
	deleteStmt    *sql.Stmt
	marshalFormat MarshalFormat
}

// Set stores the given value for the given key.
// Values are automatically marshalled to JSON or gob (depending on the configuration).
// The length of the key must not exceed 255 characters.
// The key must not be "" and the value must not be nil.
func (c Client) Set(k string, v interface{}) error {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return err
	}

	// First turn the passed object into something that MySQL can handle
	var data []byte
	var err error
	switch c.marshalFormat {
	case JSON:
		data, err = util.ToJSON(v)
	case Gob:
		data, err = util.ToGob(v)
	default:
		err = errors.New("The store seems to be configured with a marshal format that's not implemented yet")
	}
	if err != nil {
		return err
	}

	if err != nil {
		return err
	}
	_, err = c.insertStmt.Exec(k, data)
	if err != nil {
		return err
	}

	return nil
}

// Get retrieves the stored value for the given key.
// You need to pass a pointer to the value, so in case of a struct
// the automatic unmarshalling can populate the fields of the object
// that v points to with the values of the retrieved object's values.
// The length of the key must not exceed 255 characters.
// If no value is found it returns (false, nil).
// The key must not be "" and the pointer must not be nil.
func (c Client) Get(k string, v interface{}) (found bool, err error) {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return false, err
	}

	// TODO: Consider using RawBytes.
	dataPtr := new([]byte)
	err = c.getStmt.QueryRow(k).Scan(dataPtr)
	// If no value was found return false
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, err
	}
	data := *dataPtr

	switch c.marshalFormat {
	case JSON:
		return true, util.FromJSON(data, v)
	case Gob:
		return true, util.FromGob(data, v)
	default:
		return true, errors.New("The store seems to be configured with a marshal format that's not implemented yet")
	}
}

// Delete deletes the stored value for the given key.
// Deleting a non-existing key-value pair does NOT lead to an error.
// The length of the key must not exceed 255 characters.
// The key must not be "".
func (c Client) Delete(k string) error {
	if err := util.CheckKey(k); err != nil {
		return err
	}

	_, err := c.deleteStmt.Exec(k)
	return err
}

// Close closes the client.
// It must be called to return all open connections to the connection pool and to release any open resources.
func (c Client) Close() error {
	return c.c.Close()
}

// MarshalFormat is an enum for the available (un-)marshal formats of this gokv.Store implementation.
type MarshalFormat int

const (
	// JSON is the MarshalFormat for (un-)marshalling to/from JSON
	JSON MarshalFormat = iota
	// Gob is the MarshalFormat for (un-)marshalling to/from gob
	Gob
)

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
	// (Un-)marshal format.
	// Optional (JSON by default).
	MarshalFormat MarshalFormat
}

// DefaultOptions is an Options object with default values.
// DataSourceName: "root@/gokv", TableName: "Item", MaxOpenConnections: 100, MarshalFormat: JSON
var DefaultOptions = Options{
	DataSourceName:     "root@/" + defaultDBname,
	TableName:          "Item",
	MaxOpenConnections: 100,
	// No need to set MarshalFormat to JSON because its zero value is fine.
}

// NewClient creates a new MySQL client.
func NewClient(options Options) (Client, error) {
	result := Client{}

	// Set default values
	if options.DataSourceName == "" {
		options.DataSourceName = DefaultOptions.DataSourceName
	}
	if options.TableName == "" {
		options.TableName = DefaultOptions.TableName
	}
	if options.MaxOpenConnections == 0 {
		options.MaxOpenConnections = DefaultOptions.MaxOpenConnections
	} else if options.MaxOpenConnections == -1 {
		options.MaxOpenConnections = 0 // 0 actually leads to the MySQL driver using no connection limit.
	}

	db, err := sql.Open("mysql", options.DataSourceName)
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
				// We can't use the existing db object, because that would lead to the same error again.
				// So create a new one without database name, but "backup" the user provided database name,
				// which we need to create the database.
				userProvidedDBname := cfg.DBName
				cfg.DBName = ""
				dsnWithoutDBname := cfg.FormatDSN()
				tempDB, err := sql.Open("mysql", dsnWithoutDBname)
				// This temporary DB must be closed.
				defer tempDB.Close()
				if err != nil {
					return result, err
				}
				err = tempDB.Ping()
				if err != nil {
					return result, err
				}
				// No need to check if userProvidedDBname == "", because in that case the error wouldn't be 1049 (unknown database).
				// In case the user doesn't have the permission to create a database, an error is returned.
				err = createDB(tempDB, userProvidedDBname)
				if err != nil {
					return result, err
				}
				// Now the initial ping should work.
				err = db.Ping()
				if err != nil {
					return result, err
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
		err = createDB(db, defaultDBname)
		if err != nil {
			return result, err
		}
		// Also, we must replace the current value of the db pointer by the new db,
		// because calling "USE" on a database only works for the single connection that's used
		// instead of for all connections in the pool.
		db.Close()
		cfg.DBName = defaultDBname
		dsnWithDBname := cfg.FormatDSN()
		db, err = sql.Open("mysql", dsnWithDBname)
		if err != nil {
			return result, err
		}
		err = db.Ping()
		if err != nil {
			return result, err
		}
	}

	// Limit number of concurrent connections. Typical max connections on a MySQL server is 100.
	// This prevents "Error 1040: Too many connections", which otherwise occurs for examaple with 500 concurrent goroutines.
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
	insertStmt, err := db.Prepare("INSERT INTO " + options.TableName + " (k, v) VALUES (?, ?) ON DUPLICATE KEY UPDATE v = VALUES(v)")
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

	result.c = db
	result.insertStmt = insertStmt
	result.getStmt = getStmt
	result.deleteStmt = deleteStmt
	result.marshalFormat = options.MarshalFormat

	return result, nil
}

// createDB creates a database with the given name.
// Note 1: When the DataSourceName already contained a database name
// but it doesn't exist yet (error 1049 occurred during Ping()),
// the same error will occur when trying to create the database.
// So this method is only useful when the DataSourceName did NOT contain a database name.
// Note 2: Prepared statements cannot be used for creating and using databases,
// so you must make sure that dbName doesn't contain SQL injections.
func createDB(db *sql.DB, dbName string) error {
	_, err := db.Exec("CREATE DATABASE IF NOT EXISTS " + dbName)
	if err != nil {
		return err
	}
	return nil
}
