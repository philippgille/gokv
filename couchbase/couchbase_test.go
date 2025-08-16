package couchbase_test

import (
	"testing"

	"github.com/philippgille/gokv/couchbase"
	"github.com/philippgille/gokv/encoding"
	"github.com/philippgille/gokv/test"
)

func TestClient(t *testing.T) {
	connStr := `couchbase://localhost`

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
		BucketName:       "test",
		Username:         "administrator",
		Password:         "password",
	}

	return couchbase.NewClient(options)
}
