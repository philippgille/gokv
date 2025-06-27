package pgx

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/philippgille/gokv/encoding"
	"github.com/philippgille/gokv/util"
)

// Options are the options for the PostgreSQL client using pgx.
type Options struct {
	// Pool is pgxpool.Pool.
	Pool *pgxpool.Pool
	// TableName is the name of the table where the data is stored.
	TableName string
	// Codec is the encoding codec for (de)serializing the values.
	Codec encoding.Codec
}

var defaultOptions = Options{
	TableName: "Item",
	Codec:     encoding.JSON,
}

// Client is a gokv.Store implementation for PostgreSQL using pgx.
type Client struct {
	pool      *pgxpool.Pool
	codec     encoding.Codec
	tableName string

	// pgx prepares statements for us, so we don't need to do it manually
	// https://github.com/jackc/pgx/issues/791#issuecomment-660486444
	upsertStmt string
	getStmt    string
	deleteStmt string
}

// NewClient creates a new PostgreSQL client using pgx.
//
// You must call the Close() method on the client when you're done working with
// it. Alternatively, you can use the Options Pool's Close() method when you're
// done working with the connection pool.
func NewClient(options Options) (*Client, error) {
	if options.Pool == nil {
		return nil, errors.New("the Pool in the options must not be nil")
	}

	if options.TableName == "" {
		options.TableName = defaultOptions.TableName
	}
	if options.Codec == nil {
		options.Codec = defaultOptions.Codec
	}

	client := &Client{
		pool:       options.Pool,
		codec:      options.Codec,
		tableName:  options.TableName,
		upsertStmt: fmt.Sprintf("INSERT INTO %s (k, v) VALUES ($1, $2) ON CONFLICT (k) DO UPDATE SET v = EXCLUDED.v", options.TableName),
		getStmt:    fmt.Sprintf("SELECT v FROM %s WHERE k=$1", options.TableName),
		deleteStmt: fmt.Sprintf("DELETE FROM %s WHERE k=$1", options.TableName),
	}

	// Create table if it doesn't exist yet
	_, err := client.pool.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS `+client.tableName+` (k TEXT PRIMARY KEY, v BYTEA NOT NULL)`)
	if err != nil {
		_ = client.Close()
		return nil, err
	}

	return client, nil
}

// Set stores the given value for the given key.
func (c *Client) Set(k string, v any) error {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return err
	}

	data, err := c.codec.Marshal(v)
	if err != nil {
		return err
	}

	_, err = c.pool.Exec(context.Background(), c.upsertStmt, k, data)
	return err
}

// Get retrieves the stored value for the given key.
func (c *Client) Get(k string, v any) (bool, error) {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return false, err
	}

	var data []byte
	err := c.pool.QueryRow(context.Background(), c.getStmt, k).Scan(&data)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return true, c.codec.Unmarshal(data, v)
}

// Delete deletes the stored value for the given key.
func (c *Client) Delete(k string) error {
	if err := util.CheckKey(k); err != nil {
		return err
	}

	_, err := c.pool.Exec(context.Background(), c.deleteStmt, k)
	return err
}

// Close closes the connection pool.
func (c *Client) Close() error {
	c.pool.Close()
	return nil
}
