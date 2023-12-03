package tablestore_test

import (
	"errors"
	"fmt"
	"os"
	"testing"

	alitablestore "github.com/aliyun/aliyun-tablestore-go-sdk/tablestore"

	"github.com/philippgille/gokv/encoding"
	"github.com/philippgille/gokv/tablestore"
	"github.com/philippgille/gokv/test"
)

// TestClient tests if reading from, writing to and deleting from the store works properly.
// A struct is used as value. See TestTypes() for a test that is simpler but tests all types.
//
// Note: This test is only executed if the initial connection to Table Store works.
func TestClient(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to Table Store could be established. Probably not running in a proper test environment.")
	}

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
//
// Note: This test is only executed if the initial connection to Table Store works.
func TestTypes(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to Table Store could be established. Probably not running in a proper test environment.")
	}

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

// TestClientConcurrent launches a bunch of goroutines that concurrently work with the Table Store client.
//
// Note: This test is only executed if the initial connection to Table Store works.
func TestClientConcurrent(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to Table Store could be established. Probably not running in a proper test environment.")
	}

	client := createClient(t, encoding.JSON)

	// Alibaba Cloud doesn't have a "high performance" Table Store in Frankfurt (only "capacity"),
	// so we need to use the one in London, which has a higher latency,
	// and with throttling due to limited read/write capacity this inevitebly leads to timeouts
	// when starting 1000 requests concurrently.
	// I tested 1000 concurrent requests with higher timeout though and it worked fine.
	goroutineCount := 100

	test.TestConcurrentInteractions(t, goroutineCount, client)
}

// TestErrors tests some error cases.
//
// Note: This test is only executed if the initial connection to Table Store works.
func TestErrors(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to Table Store could be established. Probably not running in a proper test environment.")
	}

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
}

// TestNil tests the behaviour when passing nil or pointers to nil values to some methods.
//
// Note: This test is only executed if the initial connection to Table Store works.
func TestNil(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to Table Store could be established. Probably not running in a proper test environment.")
	}

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
//
// Note: This test is only executed if the initial connection to Table Store works.
func TestClose(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to Table Store could be established. Probably not running in a proper test environment.")
	}

	client := createClient(t, encoding.JSON)
	err := client.Close()
	if err != nil {
		t.Error(err)
	}
}

// checkConnection returns true if a connection could be made, false otherwise.
func checkConnection() bool {
	accessKeyID, found := os.LookupEnv("ALIBABA_CLOUD_TABLE_STORE_ACCESS_KEY_ID")
	if !found {
		fmt.Println("No access key ID found in the environment variable")
		return false
	}
	accessKeySecret, found := os.LookupEnv("ALIBABA_CLOUD_TABLE_STORE_ACCESS_KEY_SECRET")
	if !found {
		fmt.Println("No access key secret found in the environment variable")
		return false
	}

	client := alitablestore.NewClient("https://gokv.eu-west-1.ots.aliyuncs.com", "gokv", accessKeyID, accessKeySecret)

	_, err := client.ListTable()
	if err != nil {
		fmt.Println("Error during client.ListTable(): ", err.Error())
		return false
	}

	return true
}

func createClient(t *testing.T, codec encoding.Codec) tablestore.Client {
	accessKeyID, found := os.LookupEnv("ALIBABA_CLOUD_TABLE_STORE_ACCESS_KEY_ID")
	if !found {
		t.Fatal(errors.New("No access key ID found in the environment variable"))
	}
	accessKeySecret, found := os.LookupEnv("ALIBABA_CLOUD_TABLE_STORE_ACCESS_KEY_SECRET")
	if !found {
		t.Fatal(errors.New("No access key secret found in the environment variable"))
	}

	options := tablestore.Options{
		AccessKeyID:     accessKeyID,
		AccessKeySecret: accessKeySecret,
		EndpointURL:     "https://gokv.eu-west-1.ots.aliyuncs.com",
		InstanceName:    "gokv",
		Codec:           codec,
	}
	client, err := tablestore.NewClient(options)
	if err != nil {
		t.Fatal(err)
	}

	return client
}
