package s3

import (
	"bytes"
	"errors"
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	awss3 "github.com/aws/aws-sdk-go/service/s3"

	"github.com/philippgille/gokv/util"
)

// Client is a gokv.Store implementation for S3.
type Client struct {
	c             *awss3.S3
	bucketName    string
	marshalFormat MarshalFormat
}

// Set stores the given value for the given key.
// Values are automatically marshalled to JSON or gob (depending on the configuration).
// The key must not be "" and the value must not be nil.
func (c Client) Set(k string, v interface{}) error {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return err
	}

	// First turn the passed object into something that S3 can handle.
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

	pubObjectInput := awss3.PutObjectInput{
		Body:   bytes.NewReader(data),
		Bucket: &c.bucketName,
		Key:    &k,
	}
	_, err = c.c.PutObject(&pubObjectInput)
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
func (c Client) Get(k string, v interface{}) (found bool, err error) {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return false, err
	}

	getObjectInput := awss3.GetObjectInput{
		Bucket: &c.bucketName,
		Key:    &k,
	}
	getObjectOutput, err := c.c.GetObject(&getObjectInput)
	if err != nil {
		aerr, ok := err.(awserr.Error)
		if ok && aerr.Code() == awss3.ErrCodeNoSuchKey {
			return false, nil
		}
		return false, err
	}
	if getObjectOutput.Body == nil {
		// Return false if there's no value
		// TODO: Maybe return an error? Behaviour should be consistent across all implementations.
		return false, nil
	}
	data, err := ioutil.ReadAll(getObjectOutput.Body)
	if err != nil {
		return true, err
	}

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

	deleteObjectInput := awss3.DeleteObjectInput{
		Bucket: &c.bucketName,
		Key:    &k,
	}
	_, err := c.c.DeleteObject(&deleteObjectInput)
	return err
}

// Close closes the client.
// In the S3 implementation this doesn't have any effect.
func (c Client) Close() error {
	return nil
}

// MarshalFormat is an enum for the available (un-)marshal formats of this gokv.Store implementation.
type MarshalFormat int

const (
	// JSON is the MarshalFormat for (un-)marshalling to/from JSON
	JSON MarshalFormat = iota
	// Gob is the MarshalFormat for (un-)marshalling to/from gob
	Gob
)

// Options are the options for the S3 client.
type Options struct {
	// Name of the S3 bucket.
	BucketName string
	// Region of the S3 service you want to use.
	// Valid values: https://docs.aws.amazon.com/general/latest/gr/rande.html#ddb_region.
	// E.g. "us-west-2".
	// Optional (read from shared config file or environment variable if not set).
	// Environment variable: "AWS_REGION".
	Region string
	// AWS access key ID (part of the credentials).
	// Optional (read from shared credentials file or environment variable if not set).
	// Environment variable: "AWS_ACCESS_KEY_ID".
	AWSaccessKeyID string
	// AWS secret access key (part of the credentials).
	// Optional (read from shared credentials file or environment variable if not set).
	// Environment variable: "AWS_SECRET_ACCESS_KEY".
	AWSsecretAccessKey string
	// CustomEndpoint allows you to set a custom S3 service endpoint.
	// This is especially useful if you're running a "S3 local" Docker container for local testing.
	// Typical value for the Docker container: "http://localhost:8000".
	// See https://hub.docker.com/r/amazon/s3-local/.
	// Optional ("" by default)
	CustomEndpoint string
	// (Un-)marshal format.
	// Optional (JSON by default).
	MarshalFormat MarshalFormat
}

// DefaultOptions is an Options object with default values.
// Region: "" (use shared config file or environment variable),
// AWSaccessKeyID: "" (use shared credentials file or environment variable),
// AWSsecretAccessKey: "" (use shared credentials file or environment variable),
// CustomEndpoint: "", MarshalFormat: JSON
var DefaultOptions = Options{
	// No need to set Region, AWSaccessKeyID, AWSsecretAccessKey
	// MarshalFormat or CustomEndpoint because their Go zero values are fine.
}

// NewClient creates a new S3 client.
//
// Credentials can be set in the options, but it's recommended to either use the shared credentials file
// (Linux: "~/.aws/credentials", Windows: "%UserProfile%\.aws\credentials")
// or environment variables (AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY).
// See https://github.com/awsdocs/aws-go-developer-guide/blob/0ae5712d120d43867cf81de875cb7505f62f2d71/doc_source/configuring-sdk.rst#specifying-credentials.
func NewClient(options Options) (Client, error) {
	result := Client{}

	// Precondition check
	if options.BucketName == "" {
		return result, errors.New("The BucketName in the options must not be empty")
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
	svc := awss3.New(session)

	// Try to create bucket, even if it exists (in which case this serves as connection test).
	createBucketInput := awss3.CreateBucketInput{
		Bucket: aws.String(options.BucketName),
	}
	_, err = svc.CreateBucket(&createBucketInput)
	if err != nil {
		aerr, ok := err.(awserr.Error)
		if !ok || aerr.Code() != awss3.ErrCodeBucketAlreadyOwnedByYou {
			return result, err
		}
	}

	result.c = svc
	result.bucketName = options.BucketName
	result.marshalFormat = options.MarshalFormat

	return result, nil
}
