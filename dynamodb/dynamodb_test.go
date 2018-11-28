package dynamodb_test

import (
	"context"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	awsdynamodb "github.com/aws/aws-sdk-go/service/dynamodb"

	"github.com/philippgille/gokv/dynamodb"
	"github.com/philippgille/gokv/test"
)

// For "DynamoDB local" Docker container.
// See https://hub.docker.com/r/amazon/dynamodb-local/.
var customEndpoint = "http://localhost:8000"

// TestClient tests if reading from, writing to and deleting from the store works properly.
// A struct is used as value. See TestTypes() for a test that is simpler but tests all types.
//
// Note: This test is only executed if the initial connection to DynamoDB works.
func TestClient(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to DynamoDB could be established. Probably not running in a proper test environment.")
	}

	// Test with JSON
	t.Run("JSON", func(t *testing.T) {
		client := createClient(t, dynamodb.JSON)
		test.TestStore(client, t)
	})

	// Test with gob
	t.Run("gob", func(t *testing.T) {
		client := createClient(t, dynamodb.Gob)
		test.TestStore(client, t)
	})
}

// TestTypes tests if setting and getting values works with all Go types.
//
// Note: This test is only executed if the initial connection to DynamoDB works.
func TestTypes(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to DynamoDB could be established. Probably not running in a proper test environment.")
	}

	// Test with JSON
	t.Run("JSON", func(t *testing.T) {
		client := createClient(t, dynamodb.JSON)
		test.TestTypes(client, t)
	})

	// Test with gob
	t.Run("gob", func(t *testing.T) {
		client := createClient(t, dynamodb.Gob)
		test.TestTypes(client, t)
	})
}

// TestClientConcurrent launches a bunch of goroutines that concurrently work with the DynamoDB client.
//
// Note: This test is only executed if the initial connection to DynamoDB works.
func TestClientConcurrent(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to DynamoDB could be established. Probably not running in a proper test environment.")
	}

	client := createClient(t, dynamodb.JSON)

	goroutineCount := 1000

	test.TestConcurrentInteractions(t, goroutineCount, client)
}

// TestErrors tests some error cases.
//
// Note: This test is only executed if the initial connection to DynamoDB works.
func TestErrors(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to DynamoDB could be established. Probably not running in a proper test environment.")
	}

	// Test with a bad MarshalFormat enum value

	client := createClient(t, dynamodb.MarshalFormat(19))
	err := client.Set("foo", "bar")
	if err == nil {
		t.Error("An error should have occurred, but didn't")
	}
	// TODO: store some value for "foo", so retrieving the value works.
	// Just the unmarshalling should fail.
	// _, err = client.Get("foo", new(string))
	// if err == nil {
	// 	t.Error("An error should have occurred, but didn't")
	// }

	// Test empty key
	err = client.Set("", "bar")
	if err == nil {
		t.Error("Expected an error")
	}
	_, err = client.Get("", new(string))
	if err == nil {
		t.Error("Expected an error")
	}
	err = client.Delete("")
	if err == nil {
		t.Error("Expected an error")
	}

	// Test client creation with bad options
	options := dynamodb.Options{
		AWSaccessKeyID: "foo",
	}
	client, err = dynamodb.NewClient(options)
	if err.Error() != "When passing credentials via options, you need to set BOTH AWSaccessKeyID AND AWSsecretAccessKey" {
		t.Error("An error was expected, but didn't occur.")
	}
	options = dynamodb.Options{
		AWSsecretAccessKey: "foo",
	}
	client, err = dynamodb.NewClient(options)
	if err.Error() != "When passing credentials via options, you need to set BOTH AWSaccessKeyID AND AWSsecretAccessKey" {
		t.Error("An error was expected, but didn't occur.")
	}
	// Bad credentials on actual AWS DynamoDB endpoint (no custom endpoint for local Docker container)
	options = dynamodb.Options{
		AWSaccessKeyID:     "foo",
		AWSsecretAccessKey: "bar",
		Region:             endpoints.UsWest2RegionID,
	}
	_, err = dynamodb.NewClient(options)
	if strings.Index(err.Error(), "UnrecognizedClientException: The security token included in the request is invalid.") != 0 {
		t.Errorf("An UnrecognizedClientException was expected, but it seems like it didn't occur. Instead, the error was: %v", err)
	}
}

// TestNil tests the behaviour when passing nil or pointers to nil values to some methods.
//
// Note: This test is only executed if the initial connection to DynamoDB works.
func TestNil(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to DynamoDB could be established. Probably not running in a proper test environment.")
	}

	// Test setting nil

	t.Run("set nil with JSON marshalling", func(t *testing.T) {
		client := createClient(t, dynamodb.JSON)
		err := client.Set("foo", nil)
		if err == nil {
			t.Error("Expected an error")
		}
	})

	t.Run("set nil with Gob marshalling", func(t *testing.T) {
		client := createClient(t, dynamodb.Gob)
		err := client.Set("foo", nil)
		if err == nil {
			t.Error("Expected an error")
		}
	})

	// Test passing nil or pointer to nil value for retrieval

	createTest := func(mf dynamodb.MarshalFormat) func(t *testing.T) {
		return func(t *testing.T) {
			client := createClient(t, mf)

			// Prep
			err := client.Set("foo", test.Foo{Bar: "baz"})
			if err != nil {
				t.Error(err)
			}

			_, err = client.Get("foo", nil) // actually nil
			if err == nil {
				t.Error("An error was expected")
			}

			var i interface{} // actually nil
			_, err = client.Get("foo", i)
			if err == nil {
				t.Error("An error was expected")
			}

			var valPtr *test.Foo // nil value
			_, err = client.Get("foo", valPtr)
			if err == nil {
				t.Error("An error was expected")
			}
		}
	}
	t.Run("get with nil / nil value parameter", createTest(dynamodb.JSON))
	t.Run("get with nil / nil value parameter", createTest(dynamodb.Gob))
}

// TestClose tests if the close method returns any errors.
//
// Note: This test is only executed if the initial connection to DynamoDB works.
func TestClose(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to DynamoDB could be established. Probably not running in a proper test environment.")
	}

	client := createClient(t, dynamodb.JSON)
	err := client.Close()
	if err != nil {
		t.Error(err)
	}
}

// checkConnection returns true if a connection could be made, false otherwise.
func checkConnection() bool {
	sess, err := session.NewSession(aws.NewConfig().WithRegion(endpoints.EuCentral1RegionID).WithEndpoint(customEndpoint))
	if err != nil {
		log.Printf("An error occurred during testing the connection to the server: %v\n", err)
		return false
	}
	svc := awsdynamodb.New(sess)

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	limit := int64(1)
	listTablesInput := awsdynamodb.ListTablesInput{
		Limit: &limit,
	}
	_, err = svc.ListTablesWithContext(timeoutCtx, &listTablesInput)
	if err != nil {
		log.Printf("An error occurred during testing the connection to the server: %v\n", err)
		return false
	}
	return true
}

func createClient(t *testing.T, mf dynamodb.MarshalFormat) dynamodb.Client {
	options := dynamodb.Options{
		Region:         endpoints.EuCentral1RegionID,
		CustomEndpoint: customEndpoint,
		MarshalFormat:  mf,
	}
	client, err := dynamodb.NewClient(options)
	if err != nil {
		t.Fatal(err)
	}
	return client
}
