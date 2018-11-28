package mysql_test

import (
	"database/sql"
	"log"
	"testing"

	_ "github.com/go-sql-driver/mysql"

	"github.com/philippgille/gokv/mysql"
	"github.com/philippgille/gokv/test"
)

// TestClient tests if reading from, writing to and deleting from the store works properly.
// A struct is used as value. See TestTypes() for a test that is simpler but tests all types.
//
// Note: This test is only executed if the initial connection to MySQL works.
func TestClient(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to MySQL could be established. Probably not running in a proper test environment.")
	}

	// Test with JSON
	t.Run("JSON", func(t *testing.T) {
		client := createClient(t, mysql.JSON)
		test.TestStore(client, t)
	})

	// Test with gob
	t.Run("gob", func(t *testing.T) {
		client := createClient(t, mysql.Gob)
		test.TestStore(client, t)
	})
}

// TestTypes tests if setting and getting values works with all Go types.
//
// Note: This test is only executed if the initial connection to MySQL works.
func TestTypes(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to MySQL could be established. Probably not running in a proper test environment.")
	}

	// Test with JSON
	t.Run("JSON", func(t *testing.T) {
		client := createClient(t, mysql.JSON)
		test.TestTypes(client, t)
	})

	// Test with gob
	t.Run("gob", func(t *testing.T) {
		client := createClient(t, mysql.Gob)
		test.TestTypes(client, t)
	})
}

// TestClientConcurrent launches a bunch of goroutines that concurrently work with the MySQL client.
//
// Note: This test is only executed if the initial connection to MySQL works.
func TestClientConcurrent(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to MySQL could be established. Probably not running in a proper test environment.")
	}

	client := createClient(t, mysql.JSON)

	goroutineCount := 1000

	test.TestConcurrentInteractions(t, goroutineCount, client)
}

// TestErrors tests some error cases.
//
// Note: This test is only executed if the initial connection to MySQL works.
func TestErrors(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to MySQL could be established. Probably not running in a proper test environment.")
	}

	// Test with a bad MarshalFormat enum value

	client := createClient(t, mysql.MarshalFormat(19))
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
// Note: This test is only executed if the initial connection to MySQL works.
func TestNil(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to MySQL could be established. Probably not running in a proper test environment.")
	}

	// Test setting nil

	t.Run("set nil with JSON marshalling", func(t *testing.T) {
		client := createClient(t, mysql.JSON)
		err := client.Set("foo", nil)
		if err == nil {
			t.Error("Expected an error")
		}
	})

	t.Run("set nil with Gob marshalling", func(t *testing.T) {
		client := createClient(t, mysql.Gob)
		err := client.Set("foo", nil)
		if err == nil {
			t.Error("Expected an error")
		}
	})

	// Test passing nil or pointer to nil value for retrieval

	createTest := func(mf mysql.MarshalFormat) func(t *testing.T) {
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
	t.Run("get with nil / nil value parameter", createTest(mysql.JSON))
	t.Run("get with nil / nil value parameter", createTest(mysql.Gob))
}

// TestDBcreation tests if the DB gets created successfully when the DSN doesn't contain one.
// The other tests call createClient, which doesn't use any DSN,
// which leads to the gokv implementation to use "root@/gokv" by default.
func TestDBcreation(t *testing.T) {
	options := mysql.Options{
		DataSourceName: "root@/",
	}
	client, err := mysql.NewClient(options)
	if err != nil {
		t.Error(err)
	}
	err = client.Set("foo", "bar")
	if err != nil {
		t.Error(err)
	}
	actualPtr := new(string)
	found, err := client.Get("foo", actualPtr)
	if !found {
		t.Error("Value not found, but should've been.")
	}
	if err != nil {
		t.Error(err)
	}
	actual := *actualPtr
	if actual != "bar" {
		t.Errorf("Expected %v, but was: %v", "bar", actual)
	}
}

// TestClose tests if the close method returns any errors.
//
// Note: This test is only executed if the initial connection to MySQL works.
func TestClose(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to MySQL could be established. Probably not running in a proper test environment.")
	}

	client := createClient(t, mysql.JSON)
	err := client.Close()
	if err != nil {
		t.Error(err)
	}
}

func TestDefaultMaxOpenConnections(t *testing.T) {
	options := mysql.Options{}
	client, err := mysql.NewClient(options)
	if err != nil {
		t.Error(err)
	}

	err = client.Set("foo", "bar")
	if err != nil {
		t.Error(err)
	}
	vPtr := new(string)
	found, err := client.Get("foo", vPtr)
	if err != nil {
		t.Error(err)
	}
	if !found {
		t.Error("A value should have been found, but wasn't.")
	}
	if *vPtr != "bar" {
		t.Errorf("Expectec %v, but was %v", "bar", *vPtr)
	}
}

// checkConnection returns true if a connection could be made, false otherwise.
func checkConnection() bool {
	db, err := sql.Open("mysql", "root@/")
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

func createClient(t *testing.T, mf mysql.MarshalFormat) mysql.Client {
	options := mysql.Options{
		MaxOpenConnections: 25, // Higher values seem to lead to issues on Travis CI
	}
	options.MarshalFormat = mf
	client, err := mysql.NewClient(options)
	if err != nil {
		t.Fatal(err)
	}
	return client
}
