package sql

import (
	"errors"

	"database/sql"

	"github.com/philippgille/gokv/util"
)

// Client is a gokv.Store implementation for SQL databases.
type Client struct {
	C             *sql.DB
	InsertStmt    *sql.Stmt
	GetStmt       *sql.Stmt
	DeleteStmt    *sql.Stmt
	MarshalFormat MarshalFormat
}

// Set stores the given value for the given key.
// Values are automatically marshalled to JSON or gob (depending on the configuration).
// The length of the key must not exceed 255 characters.
// The key must not be "" and the value must not be nil.
func (c Client) Set(k string, v interface{}) error {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return err
	}

	// First turn the passed object into something that the SQL database can handle
	var data []byte
	var err error
	switch c.MarshalFormat {
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
	_, err = c.InsertStmt.Exec(k, data)
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
	err = c.GetStmt.QueryRow(k).Scan(dataPtr)
	// If no value was found return false
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, err
	}
	data := *dataPtr

	switch c.MarshalFormat {
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

	_, err := c.DeleteStmt.Exec(k)
	return err
}

// Close closes the client.
// It must be called to return all open connections to the connection pool and to release any open resources.
func (c Client) Close() error {
	return c.C.Close()
}

// MarshalFormat is an enum for the available (un-)marshal formats of this gokv.Store implementation.
type MarshalFormat int

const (
	// JSON is the MarshalFormat for (un-)marshalling to/from JSON
	JSON MarshalFormat = iota
	// Gob is the MarshalFormat for (un-)marshalling to/from gob
	Gob
)

// CreateDB creates a database with the given name.
// Note 1: When the DataSourceName already contained a database name
// but it doesn't exist yet (error 1049 occurred during Ping()),
// the same error will occur when trying to create the database.
// So this method is only useful when the DataSourceName did NOT contain a database name.
// Note 2: Prepared statements cannot be used for creating and using databases,
// so you must make sure that dbName doesn't contain SQL injections.
func CreateDB(db *sql.DB, dbName string) error {
	_, err := db.Exec("CREATE DATABASE IF NOT EXISTS " + dbName)
	if err != nil {
		return err
	}
	return nil
}
