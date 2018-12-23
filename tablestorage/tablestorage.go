package tablestorage

import (
	"crypto/md5"
	"errors"
	"fmt"
	"strconv"

	"github.com/Azure/azure-sdk-for-go/storage"

	"github.com/philippgille/gokv/encoding"
	"github.com/philippgille/gokv/util"
)

var valAttrName = "v"

// TODO: Timeout is not documented very well,
// let's assume seconds because the Go test code sets 30 in some places
// and the documentation mentions 30 seconds as maximum timeout,
// see: https://github.com/Azure/azure-sdk-for-go/blob/7971189ecf5a584b9211f2527737f94bb979644e/storage/entity_test.go#L31
// and: https://docs.microsoft.com/en-us/rest/api/storageservices/setting-timeouts-for-table-service-operations.
// Also, 1 millisecond would timeout during the test, which doesn't happen though.
var opTimeout = uint(1)
var setupTimeout = uint(2)

// Client is a gokv.Store implementation for Table Storage.
type Client struct {
	c                    *storage.Table
	partitionKeySupplier func(k string) string
	codec                encoding.Codec
}

// Set stores the given value for the given key.
// Values are automatically marshalled to JSON or gob (depending on the configuration).
// The key must not be "" and the value must not be nil.
func (c Client) Set(k string, v interface{}) error {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return err
	}

	// First turn the passed object into something that Table Storage can handle.
	data, err := c.codec.Marshal(v)
	if err != nil {
		return err
	}

	partitionKey := c.partitionKeySupplier(k)
	entity := c.c.GetEntityReference(partitionKey, k)
	valMap := make(map[string]interface{})
	valMap[valAttrName] = data
	entity.Properties = valMap
	entityOptions := storage.EntityOptions{
		Timeout: opTimeout,
	}
	err = entity.InsertOrReplace(&entityOptions)
	if err != nil {
		return err
	}
	return nil
}

// Get retrieves the stored value for the given key.
// You need to pass a pointer to the value, so in case of a struct
// the automatic unmarshalling can populate the fields of the object
// that v points to with the values of the retrieved object's values.
// If no value is found it returns (false, nil).
// The key must not be "" and the pointer must not be nil.
func (c Client) Get(k string, v interface{}) (found bool, err error) {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return false, err
	}

	partitionKey := c.partitionKeySupplier(k)
	entity := c.c.GetEntityReference(partitionKey, k)
	getEntityOptions := storage.GetEntityOptions{
		Select: []string{valAttrName},
	}
	timeout := uint(opTimeout)
	err = entity.Get(timeout, storage.FullMetadata, &getEntityOptions)
	if err != nil {
		storageErr, ok := err.(storage.AzureStorageServiceError)
		if !ok {
			return false, err
		}
		// Handle AzureStorageServiceError.
		// Return false if the key-value pair doesn't exist.
		if storageErr.Code == "ResourceNotFound" {
			return false, nil
		}
		return false, err
	}
	retrievedVal := entity.Properties[valAttrName]
	data, ok := retrievedVal.([]byte)
	if !ok {
		return true, fmt.Errorf("The value belonging to the key was expected to be a slice of bytes, but wasn't. Key: %v", k)
	}

	return true, c.codec.Unmarshal(data, v)
}

// Delete deletes the stored value for the given key.
// Deleting a non-existing key-value pair does NOT lead to an error.
// The key must not be "".
func (c Client) Delete(k string) error {
	if err := util.CheckKey(k); err != nil {
		return err
	}

	partitionKey := c.partitionKeySupplier(k)
	entity := c.c.GetEntityReference(partitionKey, k)
	entityOptions := storage.EntityOptions{
		Timeout: opTimeout,
	}
	err := entity.Delete(true, &entityOptions)
	if err != nil {
		storageErr, ok := err.(storage.AzureStorageServiceError)
		if !ok {
			return err
		}
		// Handle AzureStorageServiceError.
		// Return false if the key-value pair doesn't exist.
		if storageErr.Code == "ResourceNotFound" {
			return nil
		}
	}

	return err
}

// Close closes the client.
// In the Table Storage implementation this doesn't have any effect.
func (c Client) Close() error {
	return nil
}

// Options are the options for the Table Storage client.
type Options struct {
	// Connection string.
	// Can be either a normal connection string like this:
	// "DefaultEndpointsProtocol=https;AccountName=foo;AccountKey=abc123==;EndpointSuffix=core.windows.net".
	// Or a "shared access signature". In this case it must not contain "BlobEndpoint", "FileEndpoint" or "QueueEndpoint".
	// Example: "TableEndpoint=https://foo.table.core.windows.net/;SharedAccessSignature=sv=2017-11-09&ss=t&srt=sco&sp=rwdlacu&se=2018-01-01T00:00:00Z&st=2018-01-02T00:00:00Z&spr=https&sig=abc123"
	ConnectionString string
	// Name of the table.
	// If the table doesn't exist yet, it's created automatically.
	// Optional ("gokv" by default).
	TableName string
	// PartitionKeySupplier is a function for supplying a "partition key" for a given key.
	//
	// The partition key is used to split the storage into logical partitions,
	// which is required to scale and evenly distribute workload across physical Table Storage partitions.
	// The Table Storage documentation suggests to use several hundred to several thousand distinct values.
	//
	// When using Table Storage as key-value store WITHOUT QUERIES though, partition keys can be as fine grained as possible,
	// up to using NO partition key at all, which leads to a new partition for each entity.
	// This only has a disadvantage when the keys that are being inserted are ordered
	// (e.g. you have 5 key-value pairs, with the keys being in the order 123, 124, 125, 126, 127),
	// because Azure then creates "range partitions", which can lead to bad scalability for inserts.
	//
	// gokv's default PartitionKeySupplier returns an empty partition key,
	// 1) for maximum scalability and 2) to not force a partition key generation algorithm on you,
	// which then leads to the key-value pairs not being accessible without using the same algorithm in the future.
	//
	// For a synthetic partition key generator you can create one using tablestorage.CreateSyntheticPartitionKeySupplier().
	// It's experimental though, so we might change or remove it in future releases.
	//
	// If you want to create your own PartitionKeySupplier, you should read the documentation about how to choose a PartitonKey.
	// See https://github.com/MicrosoftDocs/azure-docs/blob/984bcf2b5adf9340004603539d47fa94e13e4568/articles/cosmos-db/partitioning-overview.md#choosing-a-partition-key.
	// And https://docs.microsoft.com/en-us/rest/api/storageservices/designing-a-scalable-partitioning-strategy-for-azure-table-storage.
	// And https://github.com/MicrosoftDocs/azure-docs/blob/f3ffbfd3258ee1132f710cfefaf33d92c4f096f2/articles/cosmos-db/synthetic-partition-keys.md.
	//
	// When you decided on a way to supply partition keys, you should never change it,
	// because already stored key-value pairs cannot be retrieved with a partition key
	// that's different from the one that was used when storing the key-value pair.
	//
	// Optional (tablestorage.EmptyPartitionKeySupplier is used, leading to NO partition keys).
	PartitionKeySupplier func(k string) string
	// Encoding format.
	// Optional (encoding.JSON by default).
	Codec encoding.Codec
}

// DefaultOptions is an Options object with default values.
// TableName: "gokv", PartitionKeySupplier: tablestorage.EmptyPartitionKeySupplier, Codec: encoding.JSON.
var DefaultOptions = Options{
	TableName:            "gokv",
	PartitionKeySupplier: EmptyPartitionKeySupplier,
	Codec:                encoding.JSON,
}

// NewClient creates a new Table Storage client.
func NewClient(options Options) (Client, error) {
	result := Client{}

	// Precondition check
	if options.ConnectionString == "" {
		return result, errors.New("The ConnectionString of the passed options is empty")
	}

	// Set default values
	if options.TableName == "" {
		options.TableName = DefaultOptions.TableName
	}
	if options.PartitionKeySupplier == nil {
		options.PartitionKeySupplier = DefaultOptions.PartitionKeySupplier
	}
	if options.Codec == nil {
		options.Codec = DefaultOptions.Codec
	}

	storageClient, err := storage.NewClientFromConnectionString(options.ConnectionString)
	if err != nil {
		return result, nil
	}

	tableService := storageClient.GetTableService()
	tableServicePtr := &tableService
	table := tableServicePtr.GetTableReference(options.TableName)
	err = table.Get(setupTimeout, storage.NoMetadata)
	if err != nil {
		storageErr, ok := err.(storage.AzureStorageServiceError)
		if !ok {
			return result, err
		}
		// Handle AzureStorageServiceError.
		// If the table wasn't found, create it.
		if storageErr.Code == "ResourceNotFound" {
			err = table.Create(setupTimeout, storage.EmptyPayload, nil)
			if err != nil {
				return result, err
			}
		} else {
			return result, err
		}
	}

	result.c = table
	result.partitionKeySupplier = options.PartitionKeySupplier
	result.codec = options.Codec

	return result, nil
}

// EmptyPartitionKeySupplier returns an empty string as partition key for any given value.
func EmptyPartitionKeySupplier(_ string) string {
	return ""
}

// CreateSyntheticPartitionKeySupplier creates a PartitionKeySupplier
// that calculates synthetic "partition keys" from the given key-value pair keys.
//
// Attention! It's experimental, so we might change or remove this in future releases.
//
// The amount of distinct partition keys can be configured via the partitionCount parameter.
// They're evenly distributed to avoid "hot spot" partitions.
// Automatic tests assert a hit count deviation below 20% across all partition keys.
// It's a "synthetic" partition key, similar to the description in the official documentation:
// https://github.com/MicrosoftDocs/azure-docs/blob/f3ffbfd3258ee1132f710cfefaf33d92c4f096f2/articles/cosmos-db/synthetic-partition-keys.md.
//
// The documentation suggests to use several hundred to several thousand distinct values.
// A high value ensures scalability and evenly distributed workload across the physical storage partitions.
// Maximum value (as enforced by uint16) is 65535.
func CreateSyntheticPartitionKeySupplier(partitionCount uint16) func(string) string {
	return func(k string) string {
		// MD5 is *cryptographically* broken, but it's still evenly distributed.
		md5Hash := md5.Sum([]byte(k))
		// An MD5 hash is 16 bytes long.
		// We now need to turn this into <partitionCount> evenly distributed strings.
		// We take the first couple of bytes and turn them into a number.
		// Then we use the modulo operation to get the exact count of distinct numbers we need.
		// To get an even distribution, the number that's devided should be a potentially much higher number
		// than the count of distinct numbers we need.
		//
		// For example:
		// If 1024 key strings are hashed with MD5,
		// and we only take the first byte of the hash and turn it into a number,
		// this would lead to 256 distinct numbers (1 byte = 8 bits)
		// and each number would occur 4 times on average.
		// Now if we use the modulo operation to turn 256 distinct numbers into 10 for example,
		// this would lead to an almost even distribution.
		// To be precise: 6 of 256 hashes (numbers 250-255) would be unevenly distributed,
		// so for example, with 1024 hashes of key strings,
		// the numbers 0-5 would have on average 104 hits, while 6-9 would have 100 hits.
		// Even with more hashes, the percentage would stay the same.
		//
		// In automatic tests it turned out that 2 bytes aren't enough for an even distribution
		// when using a partitionKeyCount of 10000 or 60000.
		// Maybe the reason was the "small" sample size of 12345678 and 123456789 hashes,
		// or it was bad luck in the MD5 randomness - I didn't do the math.
		// With 4 bytes the distribution improved and it fits into a uint64.
		i1 := uint64(md5Hash[0])
		i2 := uint64(md5Hash[1])
		i3 := uint64(md5Hash[2])
		i4 := uint64(md5Hash[3])
		i := (i1 << 24) + (i2 << 16) + (i3 << 8) + i4

		// Now i is between 0 and (1.844674407370955e+19 - 1), depending on the input, always reproducable.
		// We still need to reduce the number to <partitionCount> while keeping the even distribution.
		// Remainder calculation leads exactly to a number that's between 0 and (partitionCount-1).
		rem := i % uint64(partitionCount)

		return strconv.FormatUint(uint64(rem), 10)
	}
}
