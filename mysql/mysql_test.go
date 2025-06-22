package mysql_test

import (
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"

	"github.com/philippgille/gokv/encoding"
	"github.com/philippgille/gokv/mysql"
	"github.com/philippgille/gokv/test"
)

// TestClient tests if reading from, writing to and deleting from the store works properly.
// A struct is used as value. See TestTypes() for a test that is simpler but tests all types.
func TestClient(t *testing.T) {
	// For some reason this test fails in GitHub Actions, but not locally.
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		t.Skip("Skipping test in GitHub Actions. Run this locally before a release!")
	}

	// Test with JSON
	t.Run("JSON", func(t *testing.T) {
		client := createClient(t, encoding.JSON)
		defer client.Close()
		test.TestStore(client, t)
	})

	// Test with gob
	t.Run("gob", func(t *testing.T) {
		client := createClient(t, encoding.Gob)
		defer client.Close()
		test.TestStore(client, t)
	})
}

// TestTypes tests if setting and getting values works with all Go types.
func TestTypes(t *testing.T) {
	// For some reason this test fails in GitHub Actions, but not locally.
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		t.Skip("Skipping test in GitHub Actions. Run this locally before a release!")
	}

	// Test with JSON
	t.Run("JSON", func(t *testing.T) {
		client := createClient(t, encoding.JSON)
		defer client.Close()
		test.TestTypes(client, t)
	})

	// Test with gob
	t.Run("gob", func(t *testing.T) {
		client := createClient(t, encoding.Gob)
		defer client.Close()
		test.TestTypes(client, t)
	})
}

// TestClientConcurrent launches a bunch of goroutines that concurrently work with the MySQL client.
func TestClientConcurrent(t *testing.T) {
	// For some reason this test fails in GitHub Actions, but not locally.
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		t.Skip("Skipping test in GitHub Actions. Run this locally before a release!")
	}
	client := createClient(t, encoding.JSON)
	defer client.Close()

	goroutineCount := 1000

	test.TestConcurrentInteractions(t, goroutineCount, client)
}

// TestErrors tests some error cases.
func TestErrors(t *testing.T) {
	// For some reason this test fails in GitHub Actions, but not locally.
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		t.Skip("Skipping test in GitHub Actions. Run this locally before a release!")
	}

	// Test empty key
	client := createClient(t, encoding.JSON)
	defer client.Close()
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
}

// TestNil tests the behaviour when passing nil or pointers to nil values to some methods.
func TestNil(t *testing.T) {
	// For some reason this test fails in GitHub Actions, but not locally.
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		t.Skip("Skipping test in GitHub Actions. Run this locally before a release!")
	}

	// Test setting nil

	t.Run("set nil with JSON marshalling", func(t *testing.T) {
		client := createClient(t, encoding.JSON)
		defer client.Close()
		err := client.Set("foo", nil)
		if err == nil {
			t.Error("Expected an error")
		}
	})

	t.Run("set nil with Gob marshalling", func(t *testing.T) {
		client := createClient(t, encoding.Gob)
		defer client.Close()
		err := client.Set("foo", nil)
		if err == nil {
			t.Error("Expected an error")
		}
	})

	// Test passing nil or pointer to nil value for retrieval

	createTest := func(codec encoding.Codec) func(t *testing.T) {
		return func(t *testing.T) {
			client := createClient(t, codec)
			defer client.Close()

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

// TestDBcreation tests if the DB gets created successfully when the DSN doesn't contain one.
// The other tests call createClient, which doesn't use any DSN,
// which leads to the gokv implementation to use "root@/gokv" by default.
func TestDBcreation(t *testing.T) {
	// For some reason this test fails in GitHub Actions, but not locally.
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		t.Skip("Skipping test in GitHub Actions. Run this locally before a release!")
	}

	options := mysql.Options{
		DataSourceName: "root@/",
	}
	client, err := mysql.NewClient(options)
	if err != nil {
		t.Error(err)
	}
	defer client.Close()
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
func TestClose(t *testing.T) {
	// For some reason this test fails in GitHub Actions, but not locally.
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		t.Skip("Skipping test in GitHub Actions. Run this locally before a release!")
	}

	client := createClient(t, encoding.JSON)
	err := client.Close()
	if err != nil {
		t.Error(err)
	}
}

func TestDefaultMaxOpenConnections(t *testing.T) {
	// For some reason this test fails in GitHub Actions, but not locally.
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		t.Skip("Skipping test in GitHub Actions. Run this locally before a release!")
	}

	options := mysql.Options{}
	client, err := mysql.NewClient(options)
	defer client.Close()
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

func createClient(t *testing.T, codec encoding.Codec) mysql.Client {
	options := mysql.Options{
		Codec: codec,
		// Higher values seem to lead to issues on Travis CI
		MaxOpenConnections: 25,
	}
	client, err := mysql.NewClient(options)
	if err != nil {
		t.Fatal(err)
	}
	return client
}
