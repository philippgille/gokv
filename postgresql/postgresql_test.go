package postgresql_test

import (
	"database/sql"
	"log"
	"testing"

	_ "github.com/lib/pq"

	"github.com/philippgille/gokv/postgresql"
	"github.com/philippgille/gokv/test"
)

// TestClient tests if reading from, writing to and deleting from the store works properly.
// A struct is used as value. See TestTypes() for a test that is simpler but tests all types.
//
// Note: This test is only executed if the initial connection to PostgreSQL works.
func TestClient(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to PostgreSQL could be established. Probably not running in a proper test environment.")
	}

	// Test with JSON
	t.Run("JSON", func(t *testing.T) {
		client := createClient(t, postgresql.JSON)
		test.TestStore(client, t)
	})

	// Test with gob
	t.Run("gob", func(t *testing.T) {
		client := createClient(t, postgresql.Gob)
		test.TestStore(client, t)
	})
}

// TestTypes tests if setting and getting values works with all Go types.
//
// Note: This test is only executed if the initial connection to PostgreSQL works.
func TestTypes(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to PostgreSQL could be established. Probably not running in a proper test environment.")
	}

	// Test with JSON
	t.Run("JSON", func(t *testing.T) {
		client := createClient(t, postgresql.JSON)
		test.TestTypes(client, t)
	})

	// Test with gob
	t.Run("gob", func(t *testing.T) {
		client := createClient(t, postgresql.Gob)
		test.TestTypes(client, t)
	})
}

// TestClientConcurrent launches a bunch of goroutines that concurrently work with the PostgreSQL client.
//
// Note: This test is only executed if the initial connection to PostgreSQL works.
func TestClientConcurrent(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to PostgreSQL could be established. Probably not running in a proper test environment.")
	}

	client := createClient(t, postgresql.JSON)

	goroutineCount := 1000

	test.TestConcurrentInteractions(t, goroutineCount, client)
}

// TestErrors tests some error cases.
//
// Note: This test is only executed if the initial connection to PostgreSQL works.
func TestErrors(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to PostgreSQL could be established. Probably not running in a proper test environment.")
	}

	// Test with a bad MarshalFormat enum value

	client := createClient(t, postgresql.MarshalFormat(19))
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
}

// TestNil tests the behaviour when passing nil or pointers to nil values to some methods.
//
// Note: This test is only executed if the initial connection to PostgreSQL works.
func TestNil(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to PostgreSQL could be established. Probably not running in a proper test environment.")
	}

	// Test setting nil

	t.Run("set nil with JSON marshalling", func(t *testing.T) {
		client := createClient(t, postgresql.JSON)
		err := client.Set("foo", nil)
		if err == nil {
			t.Error("Expected an error")
		}
	})

	t.Run("set nil with Gob marshalling", func(t *testing.T) {
		client := createClient(t, postgresql.Gob)
		err := client.Set("foo", nil)
		if err == nil {
			t.Error("Expected an error")
		}
	})

	// Test passing nil or pointer to nil value for retrieval

	createTest := func(mf postgresql.MarshalFormat) func(t *testing.T) {
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
	t.Run("get with nil / nil value parameter", createTest(postgresql.JSON))
	t.Run("get with nil / nil value parameter", createTest(postgresql.Gob))
}

// TestClose tests if the close method returns any errors.
//
// Note: This test is only executed if the initial connection to PostgreSQL works.
func TestClose(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to PostgreSQL could be established. Probably not running in a proper test environment.")
	}

	client := createClient(t, postgresql.JSON)
	err := client.Close()
	if err != nil {
		t.Error(err)
	}
}

// checkConnection returns true if a connection could be made, false otherwise.
func checkConnection() bool {
	db, err := sql.Open("postgres", "postgres://postgres:secret@/?sslmode=disable")
	if err != nil {
		log.Printf("An error occurred during testing the connection to the server: %v\n", err)
		return false
	}

	err = db.Ping()
	if err != nil {
		log.Printf("An error occurred during testing the connection to the server: %v\n", err)
		return false
	}

	return true
}

func createClient(t *testing.T, mf postgresql.MarshalFormat) postgresql.Client {
	options := postgresql.Options{
		ConnectionURL: "postgres://postgres:secret@/gokv?sslmode=disable",
		MarshalFormat: mf,
		// Higher values seem to lead to issues on Travis CI when using MySQL,
		// so let's just use the same value here.
		MaxOpenConnections: 25,
	}
	client, err := postgresql.NewClient(options)
	if err != nil {
		t.Fatal(err)
	}
	return client
}
