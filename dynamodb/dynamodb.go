package dynamodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/ratelimit"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awsdynamodb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/smithy-go/ptr"

	"github.com/philippgille/gokv/encoding"
	"github.com/philippgille/gokv/util"
)

// "k" is used as table column name for the key.
var keyAttrName = "k"

// "v" is used as table column name for the value.
var valAttrName = "v"

// Client is a gokv.Store implementation for DynamoDB.
type Client struct {
	c         *awsdynamodb.Client
	tableName string
	codec     encoding.Codec
}

// Set stores the given value for the given key.
// Values are automatically marshalled to JSON or gob (depending on the configuration).
// The key must not be "" and the value must not be nil.
func (c Client) Set(k string, v any) error {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return err
	}

	// First turn the passed object into something that DynamoDB can handle.
	data, err := c.codec.Marshal(v)
	if err != nil {
		return err
	}

	item := make(map[string]types.AttributeValue)
	item[keyAttrName] = &types.AttributeValueMemberS{
		Value: k,
	}
	item[valAttrName] = &types.AttributeValueMemberB{
		Value: data,
	}
	putItemInput := awsdynamodb.PutItemInput{
		TableName: &c.tableName,
		Item:      item,
	}
	_, err = c.c.PutItem(context.Background(), &putItemInput)
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
func (c Client) Get(k string, v any) (found bool, err error) {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return false, err
	}

	key := make(map[string]types.AttributeValue)
	key[keyAttrName] = &types.AttributeValueMemberS{
		Value: k,
	}
	getItemInput := awsdynamodb.GetItemInput{
		TableName: &c.tableName,
		Key:       key,
	}
	getItemOutput, err := c.c.GetItem(context.Background(), &getItemInput)
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
	data, ok := attributeVal.(*types.AttributeValueMemberB)
	if !ok {
		return true, fmt.Errorf(`value is not string`)
	}

	return true, c.codec.Unmarshal(data.Value, v)
}

// Delete deletes the stored value for the given key.
// Deleting a non-existing key-value pair does NOT lead to an error.
// The key must not be "".
func (c Client) Delete(k string) error {
	if err := util.CheckKey(k); err != nil {
		return err
	}

	key := make(map[string]types.AttributeValue)
	key[keyAttrName] = &types.AttributeValueMemberS{
		Value: k,
	}
	deleteItemInput := awsdynamodb.DeleteItemInput{
		TableName: &c.tableName,
		Key:       key,
	}
	_, err := c.c.DeleteItem(context.Background(), &deleteItemInput)
	return err
}

// Close closes the client.
// In the DynamoDB implementation this doesn't have any effect.
func (c Client) Close() error {
	return nil
}

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
	// Encoding format.
	// Optional (encoding.JSON by default).
	Codec encoding.Codec
	// Describe Table timeout
	// Defaults to 5 * time.Second
	DescribeTimeout time.Duration
}

// DefaultOptions is an Options object with default values.
// Region: "" (use shared config file or environment variable), TableName: "gokv",
// AWSaccessKeyID: "" (use shared credentials file or environment variable),
// AWSsecretAccessKey: "" (use shared credentials file or environment variable),
// CustomEndpoint: "", Codec: encoding.JSON
var DefaultOptions = Options{
	TableName:            "gokv",
	ReadCapacityUnits:    5,
	WriteCapacityUnits:   5,
	WaitForTableCreation: ptr.Bool(true),
	Codec:                encoding.JSON,
	DescribeTimeout:      5 * time.Second,
	// No need to set Region, AWSaccessKeyID, AWSsecretAccessKey
	// or CustomEndpoint because their Go zero values are fine.
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
	if options.Codec == nil {
		options.Codec = DefaultOptions.Codec
	}
	if options.DescribeTimeout == 0 {
		options.DescribeTimeout = DefaultOptions.DescribeTimeout
	}
	// Set credentials only if set in the options.
	// If not set, the SDK uses the shared credentials file or environment variables, which is the preferred way.
	// Return an error if only one of the values is set.
	var creds credentials.StaticCredentialsProvider
	if (options.AWSaccessKeyID != "" && options.AWSsecretAccessKey == "") || (options.AWSaccessKeyID == "" && options.AWSsecretAccessKey != "") {
		return result, errors.New("when passing credentials via options, you need to set BOTH AWSaccessKeyID AND AWSsecretAccessKey")
	} else if options.AWSaccessKeyID != "" {
		// Due to the previous check we can be sure that in this case AWSsecretAccessKey is not empty as well.
		creds = credentials.NewStaticCredentialsProvider(options.AWSaccessKeyID, options.AWSsecretAccessKey, ``)
	}

	config, err := config.LoadDefaultConfig(context.Background(), config.WithRetryer(func() aws.Retryer {
		return retry.NewStandard(func(so *retry.StandardOptions) {
			so.RateLimiter = ratelimit.NewTokenRateLimit(1000000)
		})
	}))
	if err != nil {
		return result, fmt.Errorf("failed loading AWS configuration from env: %w", err)
	}
	if options.Region != "" {
		config.Region = options.Region
	}
	_, err = creds.Retrieve(context.Background())
	if err == nil {
		config.Credentials = creds
	}
	if options.CustomEndpoint != "" {
		config.BaseEndpoint = &options.CustomEndpoint
	}
	svc := awsdynamodb.NewFromConfig(config)

	// Create table if it doesn't exist.
	// Also serves as connection test.
	// Use context for timeout.
	timeoutCtx, cancel := context.WithTimeout(context.Background(), options.DescribeTimeout)
	defer cancel()
	describeTableInput := awsdynamodb.DescribeTableInput{
		TableName: &options.TableName,
	}
	_, err = svc.DescribeTable(timeoutCtx, &describeTableInput)
	if err != nil {
		var nf *types.ResourceNotFoundException
		if errors.As(err, &nf) {
			err = createTable(options.TableName, options.ReadCapacityUnits, options.WriteCapacityUnits, *options.WaitForTableCreation, describeTableInput, svc)
			if err != nil {
				return result, err
			}
		} else {
			return result, err
		}
	}

	result.c = svc
	result.tableName = options.TableName
	result.codec = options.Codec

	return result, nil
}

func createTable(tableName string, readCapacityUnits, writeCapacityUnits int64, waitForTableCreation bool, describeTableInput awsdynamodb.DescribeTableInput, svc *awsdynamodb.Client) error {
	createTableInput := awsdynamodb.CreateTableInput{
		TableName: &tableName,
		AttributeDefinitions: []types.AttributeDefinition{{
			AttributeName: &keyAttrName,
			AttributeType: types.ScalarAttributeTypeS, // For "string"
		}},
		KeySchema: []types.KeySchemaElement{{
			AttributeName: &keyAttrName,
			KeyType:       types.KeyTypeHash, // As opposed to "RANGE"
		}},
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  &readCapacityUnits,
			WriteCapacityUnits: &writeCapacityUnits,
		},
	}
	_, err := svc.CreateTable(context.Background(), &createTableInput)
	if err != nil {
		return err
	}
	// If configured (true by default), block until the table is created.
	// Typical table creation duration is 10 seconds.
	if waitForTableCreation {
		waiter := awsdynamodb.NewTableExistsWaiter(svc)
		// Wait will poll until it gets the resource status, or max wait time
		// expires.
		err := waiter.Wait(context.Background(), &describeTableInput, 15*time.Second)
		if err != nil {
			return fmt.Errorf(`the DynamoDB table couldn't be created or took too long to be created: %w`, err)
		}
	}

	return nil
}
