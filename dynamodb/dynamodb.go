package dynamodb

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	awsdynamodb "github.com/aws/aws-sdk-go/service/dynamodb"

	"github.com/philippgille/gokv/util"
)

// "k" is used as table column name for the key.
var keyAttrName = "k"

// "v" is used as table column name for the value.
var valAttrName = "v"

// Client is a gokv.Store implementation for DynamoDB.
type Client struct {
	c             *awsdynamodb.DynamoDB
	tableName     string
	marshalFormat MarshalFormat
}

// Set stores the given value for the given key.
// Values are automatically marshalled to JSON or gob (depending on the configuration).
// The key must not be "" and the value must not be nil.
func (c Client) Set(k string, v interface{}) error {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return err
	}

	// First turn the passed object into something that DynamoDB can handle.
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

	item := make(map[string]*awsdynamodb.AttributeValue)
	item[keyAttrName] = &awsdynamodb.AttributeValue{
		S: &k,
	}
	item[valAttrName] = &awsdynamodb.AttributeValue{
		B: data,
	}
	putItemInput := awsdynamodb.PutItemInput{
		TableName: &c.tableName,
		Item:      item,
	}
	_, err = c.c.PutItem(&putItemInput)
	if err != nil {
		return err
	}
	return nil
}

// Get retrieves the stored value for the given key.
// You need to pass a pointer to the value, so in case of a struct
// the automatic unmarshalling can populate the fields of the object
// that v points to with the values of the retrieved object's values.
// If no value is found it returns (false, nil).
// The key must not be "" and the pointer must not be nil.
func (c Client) Get(k string, v interface{}) (bool, error) {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return false, err
	}

	key := make(map[string]*awsdynamodb.AttributeValue)
	key[keyAttrName] = &awsdynamodb.AttributeValue{
		S: &k,
	}
	getItemInput := awsdynamodb.GetItemInput{
		TableName: &c.tableName,
		Key:       key,
	}
	getItemOutput, err := c.c.GetItem(&getItemInput)
	if err != nil {
		return false, err
	} else if getItemOutput.Item == nil {
		// Return false if the key-value pair doesn't exist
		return false, nil
	}
	attributeVal := getItemOutput.Item[valAttrName]
	if attributeVal == nil {
		// Return false if there's no value
		// TODO: Maybe return an error? Behaviour should be consistent across all implementations.
		return false, nil
	}
	data := attributeVal.B

	switch c.marshalFormat {
	case JSON:
		return true, util.FromJSON([]byte(data), v)
	case Gob:
		return true, util.FromGob([]byte(data), v)
	default:
		return true, errors.New("The store seems to be configured with a marshal format that's not implemented yet")
	}
}

// Delete deletes the stored value for the given key.
// Deleting a non-existing key-value pair does NOT lead to an error.
// The key must not be "".
func (c Client) Delete(k string) error {
	if err := util.CheckKey(k); err != nil {
		return err
	}

	key := make(map[string]*awsdynamodb.AttributeValue)
	key[keyAttrName] = &awsdynamodb.AttributeValue{
		S: &k,
	}
	deleteItemInput := awsdynamodb.DeleteItemInput{
		TableName: &c.tableName,
		Key:       key,
	}
	_, err := c.c.DeleteItem(&deleteItemInput)
	return err
}

// MarshalFormat is an enum for the available (un-)marshal formats of this gokv.Store implementation.
type MarshalFormat int

const (
	// JSON is the MarshalFormat for (un-)marshalling to/from JSON
	JSON MarshalFormat = iota
	// Gob is the MarshalFormat for (un-)marshalling to/from gob
	Gob
)

// Options are the options for the DynamoDB client.
type Options struct {
	// Region of the DynamoDB service you want to use.
	// Valid values: https://docs.aws.amazon.com/general/latest/gr/rande.html#ddb_region.
	// E.g. "us-west-2".
	// Optional (read from shared config file or environment variable if not set).
	// Environment variable: "AWS_REGION".
	Region string
	// Name of the DynamoDB table.
	// Optional ("gokv" by default).
	TableName string
	// ReadCapacityUnits of the table.
	// Only required when the table doesn't exist yet and is created by gokv.
	// Optional (5 by default, which is the same default value as when creating a table in the web console)
	// 25 RCUs are included in the free tier (across all tables).
	// For example calculations, see https://github.com/awsdocs/amazon-dynamodb-developer-guide/blob/c420420a59040c5b3dd44a6e59f7c9e55fc922ef/doc_source/HowItWorks.ProvisionedThroughput.
	// For limits, see https://github.com/awsdocs/amazon-dynamodb-developer-guide/blob/c420420a59040c5b3dd44a6e59f7c9e55fc922ef/doc_source/Limits.md#capacity-units-and-provisioned-throughput.md#provisioned-throughput.
	ReadCapacityUnits int64
	// ReadCapacityUnits of the table.
	// Only required when the table doesn't exist yet and is created by gokv.
	// Optional (5 by default, which is the same default value as when creating a table in the web console)
	// 25 RCUs are included in the free tier (across all tables).
	// For example calculations, see https://github.com/awsdocs/amazon-dynamodb-developer-guide/blob/c420420a59040c5b3dd44a6e59f7c9e55fc922ef/doc_source/HowItWorks.ProvisionedThroughput.
	// For limits, see https://github.com/awsdocs/amazon-dynamodb-developer-guide/blob/c420420a59040c5b3dd44a6e59f7c9e55fc922ef/doc_source/Limits.md#capacity-units-and-provisioned-throughput.md#provisioned-throughput.
	WriteCapacityUnits int64
	// If the table doesn't exist yet, gokv creates it.
	// If WaitForTableCreation is true, gokv will block until the table is created, with a timeout of 15 seconds.
	// If the table still doesn't exist after 15 seconds, an error is returned.
	// If WaitForTableCreation is false, gokv returns the client immediately.
	// In the latter case you need to make sure that you don't read from or write to the table before it's created,
	// because otherwise you will get ResourceNotFoundException errors.
	// Optional (true by default).
	WaitForTableCreation *bool
	// AWS access key ID (part of the credentials).
	// Optional (read from shared credentials file or environment variable if not set).
	// Environment variable: "AWS_ACCESS_KEY_ID".
	AWSaccessKeyID string
	// AWS secret access key (part of the credentials).
	// Optional (read from shared credentials file or environment variable if not set).
	// Environment variable: "AWS_SECRET_ACCESS_KEY".
	AWSsecretAccessKey string
	// CustomEndpoint allows you to set a custom DynamoDB service endpoint.
	// This is especially useful if you're running a "DynamoDB local" Docker container for local testing.
	// Typical value for the Docker container: "http://localhost:8000".
	// See https://hub.docker.com/r/amazon/dynamodb-local/.
	// Optional ("" by default)
	CustomEndpoint string
	// (Un-)marshal format.
	// Optional (JSON by default).
	MarshalFormat MarshalFormat
}

// DefaultOptions is an Options object with default values.
// Region: "" (use shared config file or environment variable), TableName: "gokv",
// AWSaccessKeyID: "" (use shared credentials file or enviroment variable),
// AWSsecretAccessKey: "" (use shared credentials file or enviroment variable),
// CustomEndpoint: "", MarshalFormat: JSON
var DefaultOptions = Options{
	TableName:            "gokv",
	ReadCapacityUnits:    5,
	WriteCapacityUnits:   5,
	WaitForTableCreation: aws.Bool(true),
	// No need to set Region, AWSaccessKeyID, AWSsecretAccessKey
	// MarshalFormat or CustomEndpoint because their Go zero values are fine.
}

// NewClient creates a new DynamoDB client.
//
// Credentials can be set in the options, but it's recommended to either use the shared credentials file
// (Linux: "~/.aws/credentials", Windows: "%UserProfile%\.aws\credentials")
// or environment variables (AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY).
// See https://github.com/awsdocs/aws-go-developer-guide/blob/0ae5712d120d43867cf81de875cb7505f62f2d71/doc_source/configuring-sdk.rst#specifying-credentials.
func NewClient(options Options) (Client, error) {
	result := Client{}

	// Set default values
	if options.TableName == "" {
		options.TableName = DefaultOptions.TableName
	}
	if options.ReadCapacityUnits == 0 {
		options.ReadCapacityUnits = DefaultOptions.ReadCapacityUnits
	}
	if options.WriteCapacityUnits == 0 {
		options.WriteCapacityUnits = DefaultOptions.WriteCapacityUnits
	}
	if options.WaitForTableCreation == nil {
		options.WaitForTableCreation = DefaultOptions.WaitForTableCreation
	}
	// Set credentials only if set in the options.
	// If not set, the SDK uses the shared credentials file or environment variables, which is the preferred way.
	// Return an error if only one of the values is set.
	var creds *credentials.Credentials
	if (options.AWSaccessKeyID != "" && options.AWSsecretAccessKey == "") || (options.AWSaccessKeyID == "" && options.AWSsecretAccessKey != "") {
		return result, errors.New("When passing credentials via options, you need to set BOTH AWSaccessKeyID AND AWSsecretAccessKey")
	} else if options.AWSaccessKeyID != "" {
		// Due to the previous check we can be sure that in this case AWSsecretAccessKey is not empty as well.
		creds = credentials.NewStaticCredentials(options.AWSaccessKeyID, options.AWSsecretAccessKey, "")
	}

	config := aws.NewConfig()
	if options.Region != "" {
		config = config.WithRegion(options.Region)
	}
	if creds != nil {
		config = config.WithCredentials(creds)
	}
	if options.CustomEndpoint != "" {
		config = config.WithEndpoint(options.CustomEndpoint)
	}
	// Use shared config file...
	sessionOpts := session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}
	// ...but allow overwrite of region and credentials if they are set in the options.
	sessionOpts.Config.MergeIn(config)
	session, err := session.NewSessionWithOptions(sessionOpts)
	if err != nil {
		return result, err
	}
	svc := awsdynamodb.New(session)

	// Create table if it doesn't exist.
	// Also serves as connection test.
	// Use context for timeout.
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	describeTableInput := awsdynamodb.DescribeTableInput{
		TableName: &options.TableName,
	}
	_, err = svc.DescribeTableWithContext(timeoutCtx, &describeTableInput)
	if err != nil {
		awsErr, ok := err.(awserr.Error)
		if !ok {
			return result, err
		} else if awsErr.Code() == awsdynamodb.ErrCodeResourceNotFoundException {
			keyAttrType := "S" // For "string"
			keyType := "HASH"  // As opposed to "RANGE"
			createTableInput := awsdynamodb.CreateTableInput{
				TableName: &options.TableName,
				AttributeDefinitions: []*awsdynamodb.AttributeDefinition{{
					AttributeName: &keyAttrName,
					AttributeType: &keyAttrType,
				}},
				KeySchema: []*awsdynamodb.KeySchemaElement{{
					AttributeName: &keyAttrName,
					KeyType:       &keyType,
				}},
				ProvisionedThroughput: &awsdynamodb.ProvisionedThroughput{
					ReadCapacityUnits:  &options.ReadCapacityUnits,
					WriteCapacityUnits: &options.WriteCapacityUnits,
				},
			}
			_, err := svc.CreateTable(&createTableInput)
			if err != nil {
				return result, err
			}
			// If configured (true by default), block until the table is created.
			// Typical table creation duration is 10 seconds.
			if *options.WaitForTableCreation {
				for try := 1; try < 16; try++ {
					describeTableOutput, err := svc.DescribeTable(&describeTableInput)
					if err != nil || *describeTableOutput.Table.TableStatus == "CREATING" {
						time.Sleep(1 * time.Second)
					}
				}
				// Last try (16th) after 15 seconds of waiting.
				// Now handle error as such.
				describeTableOutput, err := svc.DescribeTable(&describeTableInput)
				if err != nil {
					return result, errors.New("The DynamoDB table couldn't be created")
				}
				if *describeTableOutput.Table.TableStatus == "CREATING" {
					return result, errors.New("The DynamoDB table took too long to be created")
				}
			}
		} else {
			return result, err
		}
	}

	result.c = svc
	result.tableName = options.TableName
	result.marshalFormat = options.MarshalFormat

	return result, nil
}
