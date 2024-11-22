package tablestorage_test

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"testing"

	"github.com/Azure/azure-sdk-for-go/storage"

	"github.com/philippgille/gokv/encoding"
	"github.com/philippgille/gokv/tablestorage"
	"github.com/philippgille/gokv/test"
)

var connectionStringEnvVar = "TABLE_STORAGE_CONNECTION_STRING"

// TestClient tests if reading from, writing to and deleting from the store works properly.
// A struct is used as value. See TestTypes() for a test that is simpler but tests all types.
//
// Note: This test is only executed if the initial connection to Table Storage works.
func TestClient(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to Table Storage could be established. Probably not running in a proper test environment.")
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
// Note: This test is only executed if the initial connection to Table Storage works.
func TestTypes(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to Table Storage could be established. Probably not running in a proper test environment.")
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

// TestClientConcurrent launches a bunch of goroutines that concurrently work with the Table Storage client.
//
// Note: This test is only executed if the initial connection to Table Storage works.
func TestClientConcurrent(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to Table Storage could be established. Probably not running in a proper test environment.")
	}

	client := createClient(t, encoding.JSON)

	goroutineCount := 1000

	test.TestConcurrentInteractions(t, goroutineCount, client)
}

// TestErrors tests some error cases.
//
// Note: This test is only executed if the initial connection to Table Storage works.
func TestErrors(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to Table Storage could be established. Probably not running in a proper test environment.")
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

	// Test empty connection string
	options := tablestorage.Options{}
	_, err = tablestorage.NewClient(options)
	if err == nil {
		t.Error("An error was expected")
	} else if err.Error() != "The ConnectionString of the passed options is empty" {
		t.Error("A different error was expected")
	}
}

// TestNil tests the behaviour when passing nil or pointers to nil values to some methods.
//
// Note: This test is only executed if the initial connection to Table Storage works.
func TestNil(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to Table Storage could be established. Probably not running in a proper test environment.")
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
// Note: This test is only executed if the initial connection to Table Storage works.
func TestClose(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to Table Storage could be established. Probably not running in a proper test environment.")
	}

	client := createClient(t, encoding.JSON)
	err := client.Close()
	if err != nil {
		t.Error(err)
	}
}

// TestEmptyPartitionKeySupplier makes sure that the EmptyPartitionKeySupplier
// always only returns empty strings.
func TestEmptyPartitionKeySupplier(t *testing.T) {
	emptyPartitionKeySupplier := tablestorage.EmptyPartitionKeySupplier
	res := emptyPartitionKeySupplier("foo")
	res += emptyPartitionKeySupplier("bar")
	res += emptyPartitionKeySupplier("123")
	res += emptyPartitionKeySupplier("")
	res += emptyPartitionKeySupplier("loooooooooooooooooong string")
	if res != "" {
		t.Error("emptyPartitionKeySupplier returned a non-empty string")
	}
}

// TestSyntheticPartitionKeySupplier tests if the default PartitionKeySupplier
// only creates as many partition keys as given as partitionKeyCount,
// as well as if the generated keys are evenly distributed, given a set of similar keys.
func TestSyntheticPartitionKeySupplier(t *testing.T) {
	t.Skip("Takes too much CPU and time to run regularly. Enable when working on partitioning.")
	testCases := []struct {
		partitionKeyCount int
		keyCount          int
	}{
		{10, 1234567},
		{254, 1234567},
		{255, 1234567},
		{256, 1234567},
		{257, 1234567},
		{500, 1234567},
		{512, 1234567},
		{9999, 12345678},
		{10000, 12345678},
		{10001, 12345678},
		//{60000, 123456789},
	}

	keyPrefix := "foo123-"
	for _, testCase := range testCases {
		partitionKeyCount := testCase.partitionKeyCount
		keyCount := testCase.keyCount
		log.Printf("Testcase: partitionKeyCount=%v,keyCount=%v\n", partitionKeyCount, keyCount)

		partitionKeySupplier := tablestorage.CreateSyntheticPartitionKeySupplier(uint16(partitionKeyCount))
		partitionKeyMap := make(map[string]int, partitionKeyCount)

		for i := 0; i < keyCount; i++ {
			keyNo := strconv.Itoa(i)
			key := keyPrefix + keyNo
			partitionKey := partitionKeySupplier(key)
			partitionKeyMap[partitionKey] = partitionKeyMap[partitionKey] + 1
		}

		// The map should contain only <partitionKeyCount> keys
		if len(partitionKeyMap) != partitionKeyCount {
			t.Errorf("The partition key map has an invalid amount of entries. Expected: %v, but was: %v", partitionKeyCount, len(partitionKeyMap))
		}
		// The count of each entry should be roughly the same.
		// Alow a difference of +-20% between the highest and the lowest count.
		// 20% don't lead to a "hot spot" partition.
		avg := keyCount / partitionKeyCount
		min := int(float64(avg) * 0.8)
		max := int(float64(avg) * 1.2)
		for k, v := range partitionKeyMap {
			// fmt.Printf("partition key: %v, hits: %v\n", k, v)
			if v < min || v > max {
				t.Errorf("One of the partition keys deviated more than 20%% from the expected average count. Avg: %v, key: %v, hits: %v. Testcase: partitionKeyCount=%v,keyCount=%v", avg, k, v, partitionKeyCount, keyCount)
			}
		}
		// And just to be sure: Add up all partition key counts and check if it's the expected amount
		var sum int
		for _, v := range partitionKeyMap {
			sum += v
		}
		if sum != keyCount {
			t.Errorf("The amount of generated partition keys doesn't match the given / expected amount of partition keys. Testcase: %v", partitionKeyCount)
		}
	}
}

// checkConnection returns true if a connection could be made, false otherwise.
func checkConnection() bool {
	// This is the standard storage emulator connection string,
	// see https://github.com/Azure/Azurite/blob/v3.17.1/README.md#connection-strings.
	// And it leads to a regular emulator client. But the Go SDK doesn't seem to work with it. // TODO: Investigate / create GitHub issue.
	//connString := "DefaultEndpointsProtocol=http;AccountName=devstoreaccount1;AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;TableEndpoint=http://127.0.0.1:10002/devstoreaccount1;"
	// Alternatively test with a provided conn string for example for the real Azure.
	connString, found := os.LookupEnv(connectionStringEnvVar)
	if !found {
		fmt.Println("No connection string found in the environment variable")
		return false
	}
	storageClient, err := storage.NewClientFromConnectionString(connString)
	if err != nil {
		fmt.Printf("Error creating storage client from connection string: %v\n", err)
		return false
	}
	tableService := storageClient.GetTableService()
	tableServicePtr := &tableService
	_, err = tableServicePtr.QueryTables(storage.MinimalMetadata, nil)
	if err != nil {
		fmt.Printf("Error retrieving table info: %v\n", err)
		return false
	}
	return true
}

func createClient(t *testing.T, codec encoding.Codec) tablestorage.Client {
	// This is the standard storage emulator connection string,
	// see https://github.com/Azure/Azurite/blob/v3.17.1/README.md#connection-strings.
	// And it leads to a regular emulator client. But the Go SDK doesn't seem to work with it. // TODO: Investigate / create GitHub issue.
	//connString := "DefaultEndpointsProtocol=http;AccountName=devstoreaccount1;AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;TableEndpoint=http://127.0.0.1:10002/devstoreaccount1;"
	// Alternatively test with a provided conn string for example for the real Azure.
	connString, found := os.LookupEnv(connectionStringEnvVar)
	if !found {
		t.Fatal(errors.New("No connection string found in the environment variable"))
	}
	options := tablestorage.Options{
		ConnectionString: connString,
		Codec:            codec,
	}
	client, err := tablestorage.NewClient(options)
	if err != nil {
		t.Fatal(err)
	}
	return client
}
