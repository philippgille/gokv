/*
Package mongodb contains an implementation of the `gokv.Store` interface for MongoDB.

Note: If you use a sharded cluster, you must use "_id" as the shard key!
You should also use hashed sharding as opposed to ranged sharding to enable more evenly distributed data no matter how your key looks like.
*/
package mongodb
