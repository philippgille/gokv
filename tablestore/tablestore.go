package tablestore

import (
	"errors"
	"strings"
	"time"

	"github.com/aliyun/aliyun-tablestore-go-sdk/tablestore"

	"github.com/philippgille/gokv/encoding"
	"github.com/philippgille/gokv/util"
)

var keyAttrName = "k"

// Client is a gokv.Store implementation for Table Store.
type Client struct {
	c         *tablestore.TableStoreClient
	tableName string
	codec     encoding.Codec
}

// Set stores the given value for the given key.
// Values are automatically marshalled to JSON or gob (depending on the configuration).
// The key must not be "" and the value must not be nil.
func (c Client) Set(k string, v interface{}) error {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return err
	}

	// First turn the passed object into something that Table Store can handle.
	data, err := c.codec.Marshal(v)
	if err != nil {
		return err
	}

	putRowRequest := tablestore.PutRowRequest{
		PutRowChange: &tablestore.PutRowChange{
			Condition: &tablestore.RowCondition{
				RowExistenceExpectation: tablestore.RowExistenceExpectation_IGNORE,
			},
			Columns: []tablestore.AttributeColumn{tablestore.AttributeColumn{
				ColumnName: "v",
				Value:      data,
			}},
			PrimaryKey: &tablestore.PrimaryKey{
				PrimaryKeys: []*tablestore.PrimaryKeyColumn{&tablestore.PrimaryKeyColumn{
					ColumnName: keyAttrName,
					Value:      k,
				}},
			},
			TableName: c.tableName,
		},
	}
	_, err = c.c.PutRow(&putRowRequest)
	return err
}

// Get retrieves the stored value for the given key.
// You need to pass a pointer to the value, so in case of a struct
// the automatic unmarshalling can populate the fields of the object
// that v points to with the values of the retrieved object's values.
// If no value is found it returns (false, nil).
// The key must not be "" and the pointer must not be nil.
func (c Client) Get(k string, v interface{}) (found bool, err error) {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return false, err
	}

	getRowRequest := tablestore.GetRowRequest{
		SingleRowQueryCriteria: &tablestore.SingleRowQueryCriteria{
			ColumnsToGet: []string{"v"},
			MaxVersion:   1,
			PrimaryKey: &tablestore.PrimaryKey{
				PrimaryKeys: []*tablestore.PrimaryKeyColumn{&tablestore.PrimaryKeyColumn{
					ColumnName: keyAttrName,
					Value:      k,
				}},
			},
			TableName: c.tableName,
		},
	}
	getRowResponse, err := c.c.GetRow(&getRowRequest)
	if err != nil {
		return false, err
	}
	// Return false if no value was found
	if len(getRowResponse.Columns) == 0 {
		return false, nil
	} else if len(getRowResponse.Columns) > 1 {
		return false, errors.New("The returned GetRowResponse should only contain one column, but it contains more than one")
	}
	attributeColumn := getRowResponse.Columns[0]
	if attributeColumn == nil {
		return true, errors.New("A key-value pair for the key was found, but it didn't contain a value")
	}
	dataIface := attributeColumn.Value
	data, ok := dataIface.([]byte)
	if !ok {
		return false, errors.New("The returned value was expected to be a slice of bytes, but it wasn't")
	}

	return true, c.codec.Unmarshal(data, v)
}

// Delete deletes the stored value for the given key.
// Deleting a non-existing key-value pair does NOT lead to an error.
// The key must not be "".
func (c Client) Delete(k string) error {
	if err := util.CheckKey(k); err != nil {
		return err
	}

	deleteRowRequest := tablestore.DeleteRowRequest{
		DeleteRowChange: &tablestore.DeleteRowChange{
			Condition: &tablestore.RowCondition{
				RowExistenceExpectation: tablestore.RowExistenceExpectation_IGNORE,
			},
			PrimaryKey: &tablestore.PrimaryKey{
				PrimaryKeys: []*tablestore.PrimaryKeyColumn{&tablestore.PrimaryKeyColumn{
					ColumnName: keyAttrName,
					Value:      k,
				}},
			},
			TableName: c.tableName,
		},
	}
	_, err := c.c.DeleteRow(&deleteRowRequest)

	return err
}

// Close closes the client.
// In the Table Store implementation this doesn't have any effect.
func (c Client) Close() error {
	return nil
}

// Options are the options for the Table Store client.
type Options struct {
	// URL of the endpoint.
	// E.g. "https://mytable.ap-southeast-1.ots.aliyuncs.com".
	EndpointURL string
	// Name of the instance.
	// E.g. "mytable".
	InstanceName string
	// AccessKey ID.
	AccessKeyID string
	// AccessKey secret.
	AccessKeySecret string
	// Name of the table.
	// If the table doesn't exist yet, it's created automatically.
	// Optional ("gokv" by default).
	TableName string
	// Reserved read capacity.
	// 0 works fine, but doesn't *guarantee* any capacity.
	// Optional (0 by default).
	ReservedReadCap int
	// Reserved write capacity.
	// 0 works fine, but doesn't *guarantee* any capacity.
	// Optional (0 by default).
	ReservedWriteCap int
	// Encoding format.
	// Optional (encoding.JSON by default).
	Codec encoding.Codec
}

// DefaultOptions is an Options object with default values.
// TableName: "gokv", ReservedReadCap: 0, ReservedWriteCap: 0, Codec: encoding.JSON.
var DefaultOptions = Options{
	TableName: "gokv",
	Codec:     encoding.JSON,
	// No need to set ReservedReadCap or ReservedWriteCap because their Go zero values are fine.
}

// NewClient creates a new Table Store client.
func NewClient(options Options) (Client, error) {
	result := Client{}

	// Precondition check
	if options.EndpointURL == "" {
		return result, errors.New("The EndpointURL of the passed options is empty")
	} else if options.InstanceName == "" {
		return result, errors.New("The InstanceName of the passed options is empty")
	} else if options.AccessKeyID == "" {
		return result, errors.New("The AccessKeyID of the passed options is empty")
	} else if options.AccessKeySecret == "" {
		return result, errors.New("The AccessKeySecret of the passed options is empty")
	}

	// Set default values
	if options.TableName == "" {
		options.TableName = DefaultOptions.TableName
	}
	if options.Codec == nil {
		options.Codec = DefaultOptions.Codec
	}

	config := tablestore.NewDefaultTableStoreConfig()
	// Connection will stay open for multiple requests.
	// Opening the connection may take 2 seconds,
	// a request may take 1 second.
	config.HTTPTimeout.ConnectionTimeout = 2 * time.Second
	config.HTTPTimeout.RequestTimeout = time.Second
	// Default is 5 seconds and 10 retries, which is way too much
	config.MaxRetryTime = time.Second
	config.RetryTimes = 1
	client := tablestore.NewClientWithConfig(options.EndpointURL, options.InstanceName, options.AccessKeyID, options.AccessKeySecret, "", config)

	// Create table if it doesn't exist.
	// Just try to create directly and ignore the error if it's "OTSObjectAlreadyExist",
	// because this way we always have to make just one request instead of sometimes one and sometimes two
	// (one for checking if the table exists and one for creating the table).
	var primaryKeyType = tablestore.PrimaryKeyType_STRING
	createtableRequest := tablestore.CreateTableRequest{
		ReservedThroughput: &tablestore.ReservedThroughput{
			Readcap:  options.ReservedReadCap,
			Writecap: options.ReservedWriteCap,
		},
		// TODO: Are IndexMetas required when we don't need any indexes on non-primary key columns?
		TableMeta: &tablestore.TableMeta{
			DefinedColumns: []*tablestore.DefinedColumnSchema{&tablestore.DefinedColumnSchema{
				ColumnType: tablestore.DefinedColumn_BINARY,
				Name:       "v",
			}},
			SchemaEntry: []*tablestore.PrimaryKeySchema{&tablestore.PrimaryKeySchema{
				Name: &keyAttrName,
				Type: &primaryKeyType,
			}},
			TableName: options.TableName,
		},
		TableOption: &tablestore.TableOption{
			// These are the default values as set by the web interface when creating a table.
			// There they're called "Max Versions" and "Time To Live".
			MaxVersion:  1,
			TimeToAlive: -1,
		},
	}
	_, err := client.CreateTable(&createtableRequest)
	if err != nil && !strings.HasPrefix(err.Error(), "OTSObjectAlreadyExist") {
		return result, err
	}

	result.c = client
	result.tableName = options.TableName
	result.codec = options.Codec

	return result, nil
}
