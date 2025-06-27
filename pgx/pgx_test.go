package pgx_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/philippgille/gokv/encoding"
	"github.com/philippgille/gokv/pgx"
	"github.com/philippgille/gokv/test"
)

func TestClient(t *testing.T) {

	for _, tc := range []encoding.Codec{
		encoding.JSON,
		encoding.Gob,
	} {
		var codec string
		switch tc {
		case encoding.JSON:
			codec = "JSON"
		case encoding.Gob:
			codec = "gob"
		}
		t.Run(fmt.Sprintf("codec=%s", codec), func(t *testing.T) {
			t.Run("store", func(t *testing.T) {
				client := createClient(t, tc)
				defer func() { _ = client.Close() }()
				test.TestStore(client, t)
			})
			t.Run("types", func(t *testing.T) {
				client := createClient(t, tc)
				defer func() { _ = client.Close() }()
				test.TestTypes(client, t)
			})
			t.Run("concurrent interactions", func(t *testing.T) {
				client := createClient(t, tc)
				defer func() { _ = client.Close() }()
				test.TestConcurrentInteractions(t, 10, client)
			})
		})
	}
}

var i int

func createClient(t *testing.T, codec encoding.Codec) *pgx.Client {
	options := pgx.Options{
		Codec:     codec,
		TableName: fmt.Sprintf("test_table_%d", i),
	}
	i++

	p, err := pgxpool.New(context.Background(), "postgres://postgres:secret@localhost:5433/pgx?sslmode=disable")
	if err != nil {
		t.Fatalf("Unable to connect to database: %v", err)
	}
	options.Pool = p
	client, err := pgx.NewClient(options)
	if err != nil {
		t.Fatal(err)
	}
	return client
}
