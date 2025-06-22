package s3_test

import (
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws/endpoints"

	"github.com/philippgille/gokv/encoding"
	"github.com/philippgille/gokv/s3"
	"github.com/philippgille/gokv/test"
)

// For Minio Docker container.
// See https://docs.minio.io/docs/minio-docker-quickstart-guide.html.
var customEndpoint = "http://localhost:9000"

// TestClient tests if reading from, writing to and deleting from the store works properly.
// A struct is used as value. See TestTypes() for a test that is simpler but tests all types.
func TestClient(t *testing.T) {
	// Test with JSON
	t.Run("JSON", func(t *testing.T) {
		client := createClient(t, encoding.JSON)
		test.TestStore(client, t)
	})

	// Test with gob
	t.Run("gob", func(t *testing.T) {
		client := createClient(t, encoding.Gob)
		test.TestStore(client, t)
	})
}

// TestTypes tests if setting and getting values works with all Go types.
func TestTypes(t *testing.T) {
	// Test with JSON
	t.Run("JSON", func(t *testing.T) {
		client := createClient(t, encoding.JSON)
		test.TestTypes(client, t)
	})

	// Test with gob
	t.Run("gob", func(t *testing.T) {
		client := createClient(t, encoding.Gob)
		test.TestTypes(client, t)
	})
}

// TestClientConcurrent launches a bunch of goroutines that concurrently work with the S3 client.
func TestClientConcurrent(t *testing.T) {
	client := createClient(t, encoding.JSON)

	goroutineCount := 1000

	test.TestConcurrentInteractions(t, goroutineCount, client)
}

// TestErrors tests some error cases.
func TestErrors(t *testing.T) {
	// Test empty key
	client := createClient(t, encoding.JSON)
	err := client.Set("", "bar")
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
	options := s3.Options{
		AWSaccessKeyID:     "foo",
		AWSsecretAccessKey: "bar",
	}
	_, err = s3.NewClient(options)
	if err.Error() != "The BucketName in the options must not be empty" {
		t.Error("An error was expected, but didn't occur.")
	}
	options = s3.Options{
		BucketName:     "gokv",
		AWSaccessKeyID: "foo",
	}
	_, err = s3.NewClient(options)
	if err.Error() != "When passing credentials via options, you need to set BOTH AWSaccessKeyID AND AWSsecretAccessKey" {
		t.Error("An error was expected, but didn't occur.")
	}
	options = s3.Options{
		BucketName:         "gokv",
		AWSsecretAccessKey: "foo",
	}
	_, err = s3.NewClient(options)
	if err.Error() != "When passing credentials via options, you need to set BOTH AWSaccessKeyID AND AWSsecretAccessKey" {
		t.Error("An error was expected, but didn't occur.")
	}
	// Bad credentials on actual AWS S3 endpoint (no custom endpoint for local Docker container)
	options = s3.Options{
		BucketName:         "gokv",
		AWSaccessKeyID:     "foo",
		AWSsecretAccessKey: "bar",
		Region:             endpoints.UsWest2RegionID,
	}
	_, err = s3.NewClient(options)
	if strings.Index(err.Error(), "InvalidAccessKeyId: The AWS Access Key Id you provided does not exist in our records.") != 0 {
		t.Errorf("An InvalidAccessKeyId error was expected, but it seems like it didn't occur. Instead, the error was: %v", err)
	}
}

// TestNil tests the behaviour when passing nil or pointers to nil values to some methods.
func TestNil(t *testing.T) {
	// Test setting nil

	t.Run("set nil with JSON marshalling", func(t *testing.T) {
		client := createClient(t, encoding.JSON)
		err := client.Set("foo", nil)
		if err == nil {
			t.Error("Expected an error")
		}
	})

	t.Run("set nil with Gob marshalling", func(t *testing.T) {
		client := createClient(t, encoding.Gob)
		err := client.Set("foo", nil)
		if err == nil {
			t.Error("Expected an error")
		}
	})

	// Test passing nil or pointer to nil value for retrieval

	createTest := func(codec encoding.Codec) func(t *testing.T) {
		return func(t *testing.T) {
			client := createClient(t, codec)

			// Prep
			err := client.Set("foo", test.Foo{Bar: "baz"})
			if err != nil {
				t.Error(err)
			}

			_, err = client.Get("foo", nil) // actually nil
			if err == nil {
				t.Error("An error was expected")
			}

			var i any // actually nil
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
	t.Run("get with nil / nil value parameter", createTest(encoding.JSON))
	t.Run("get with nil / nil value parameter", createTest(encoding.Gob))
}

// TestClose tests if the close method returns any errors.
func TestClose(t *testing.T) {
	client := createClient(t, encoding.JSON)
	err := client.Close()
	if err != nil {
		t.Error(err)
	}
}

func createClient(t *testing.T, codec encoding.Codec) s3.Client {
	os.Setenv("AWS_ACCESS_KEY_ID", "AKIAIOSFODNN7EXAMPLE")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY")
	options := s3.Options{
		BucketName:             "gokv",
		Region:                 endpoints.EuCentral1RegionID,
		CustomEndpoint:         customEndpoint,
		UsePathStyleAddressing: true,
		Codec:                  codec,
	}
	client, err := s3.NewClient(options)
	if err != nil {
		t.Fatal(err)
	}
	return client
}
