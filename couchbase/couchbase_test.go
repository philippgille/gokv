package couchbase_test

import (
	"os"
	"testing"

	"github.com/philippgille/gokv/couchbase"
	"github.com/philippgille/gokv/encoding"
	"github.com/philippgille/gokv/test"
)

const (
	cbConnectionStringEnvVar = "COUCHBASE_SERVER"
	cbBucketNameEnvVar       = "COUCHBASE_BUCKET"
	cbUsernameEnvVar         = "COUCHBASE_USERNAME"
	cbPasswordEnvVar         = "COUCHBASE_PASSWORD"
)

func TestClient(t *testing.T) {
	t.Logf("checking for couchbase connection string on env var %q", cbConnectionStringEnvVar)

	connStr, ok := os.LookupEnv(cbConnectionStringEnvVar)
	if !ok {
		t.Skipf("unable to connect to couchbase server: env var %q not set", cbConnectionStringEnvVar)
	}

	t.Run("Test client with codec JSON", func(t *testing.T) {
		client, err := createClient(connStr, encoding.JSON)
		if err != nil {
			t.Fatalf("unable to create couchbase client: %v", err)
		}

		defer client.Close()

		test.TestStore(client, t)
	})

	t.Run("Test client with codec Gob", func(t *testing.T) {
		client, err := createClient(connStr, encoding.Gob)
		if err != nil {
			t.Fatalf("unable to create couchbase client: %v", err)
		}

		defer client.Close()

		test.TestStore(client, t)
	})
}

func createClient(connStr string, codec encoding.Codec) (*couchbase.Client, error) {
	options := couchbase.Options{
		ConnectionString: connStr,
		Codec:            codec,
		BucketName:       os.Getenv(cbBucketNameEnvVar),
		Username:         os.Getenv(cbUsernameEnvVar),
		Password:         os.Getenv(cbPasswordEnvVar),
	}

	return couchbase.NewClient(options)
}
