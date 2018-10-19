gokv
====

[![GoDoc](http://www.godoc.org/github.com/philippgille/gokv?status.svg)](http://www.godoc.org/github.com/philippgille/gokv) [![Build Status](https://travis-ci.org/philippgille/gokv.svg?branch=master)](https://travis-ci.org/philippgille/gokv) [![Go Report Card](https://goreportcard.com/badge/github.com/philippgille/gokv)](https://goreportcard.com/report/github.com/philippgille/gokv) [![GitHub Releases](https://img.shields.io/github/release/philippgille/gokv.svg)](https://github.com/philippgille/gokv/releases)

Simple key-value store abstraction and implementations for Go

Features
--------

Simple interface:

> Note: The interface is not final yet! See [Project status](#project-status) for details.

```go
// Store is an abstraction for different key-value store implementations.
// A store must be able to store and retrieve key-value pairs,
// with the key being a string and the value being any Go interface{}.
type Store interface {
	// Set stores the given value for the given key.
	// The implementation automatically marshalls the value if required.
	// The marshalling target depends on the implementation. It can be JSON, gob etc.
	// Implementations should offer a configuration for this.
	Set(string, interface{}) error
	// Get retrieves the value for the given key.
	// The implementation automatically unmarshalls the value if required.
	// The unmarshalling source depends on the implementation. It can be JSON, gob etc.
	// The automatic unmarshalling requires a pointer to a proper type being passed as parameter.
	// The Get method will populate the fields of the object that the passed pointer
	// points to with the values of the retrieved object's values.
	// If no object is found it returns (false, nil).
	Get(string, interface{}) (bool, error)
}
```

Implementations ([![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](http://makeapullrequest.com)):

- Local in-memory
    - [X] Go map (`sync.Map`)
        - Faster then a regular map when there are very few writes but lots of reads
    - [ ] Go map (`map[string]interface{}` with `sync.RWMutex`)
- Embedded
    - [X] [bbolt](https://github.com/etcd-io/bbolt) (formerly known as [Bolt / Bolt DB](https://github.com/boltdb/bolt))
        - bbolt is a fork of Bolt which was maintained by CoreOS, and now by RedHat (since CoreOS was acquired by them)
        - It's used for example in [etcd](https://github.com/etcd-io/etcd) as underlying persistent store
    - [ ] [BadgerDB](https://github.com/dgraph-io/badger)
        - Very similar to bbolt / Bolt, where bbolt is generally faster for reads, and Badger is generally faster for writes
    - [ ] [LevelDB / goleveldb](https://github.com/syndtr/goleveldb)
- Distributed
    - [X] [Redis](https://github.com/antirez/redis)
    - [ ] [Consul](https://github.com/hashicorp/consul)
    - [ ] [etcd](https://github.com/etcd-io/etcd)
        - Not advertised as general key-value store, but can be used as one
    - [ ] [Memcached](https://github.com/memcached/memcached)
    - [ ] [Hazelcast](https://github.com/hazelcast/hazelcast)
    - [ ] [LedisDB](https://github.com/siddontang/ledisdb)
        - Similar to Redis, with several backing stores
    - [ ] [TiKV](https://github.com/tikv/tikv)
        - Originally created to complement [TiDB](https://github.com/pingcap/tidb), but recently [became a project in the CNCF](https://www.cncf.io/blog/2018/08/28/cncf-to-host-tikv-in-the-sandbox/)
- Cloud
    - [ ] [Amazon DynamoDB](https://aws.amazon.com/dynamodb/)
    - [ ] [Amazon SimpleDB](https://aws.amazon.com/simpledb/)
    - [ ] [Azure Cosmos DB](https://azure.microsoft.com/en-us/services/cosmos-db/)
    - [ ] [Azure Table Storage](https://azure.microsoft.com/en-us/services/storage/tables/)
    - [ ] [Google Cloud Datastore](https://cloud.google.com/datastore/)

Some other databases aren't specifically engineered for storing key-value pairs, but if someone's running them already for other purposes and doesn't want to set up one of the proper key-value stores due to administrative overhead etc., they can of course be used as well. Let's focus on a few of the most popular though:

- SQL
    - [ ] [MySQL](https://www.mysql.com/)
    - [ ] [PostgreSQL](https://www.postgresql.org/)
- NoSQL
    - [ ] [MongoDB](https://www.mongodb.com/)
- NewSQL
    - [ ] [CockroachDB](https://github.com/cockroachdb/cockroach)
    - [ ] [TiDB](https://github.com/pingcap/tidb)

Project status
--------------

> Note: `gokv`'s API is not stable yet and is under active development. Upcoming releases are likely to contain breaking changes as long as the version is `v0.x.y`. You should use vendoring to prevent bad surprises. This project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html) and all notable changes to this project are documented in [RELEASES.md](https://github.com/philippgille/gokv/blob/master/RELEASES.md).

Planned interface methods until `v1.0.0`:

- `Delete(string) error` or similar
- `List(interface{}) error` / `GetAll(interface{}) error` or similar

Motivation
----------

Initially developed as `storage` package within the project [ln-paywall](https://github.com/philippgille/ln-paywall) to provide the users of ln-paywall with multiple storage options, at some point it made sense to turn it into a repository of its own.

Before doing so I examined existing Go packages with a similar purpose (see [Related projects](#related-projects), but none of them fit my needs. They either had too few implementations, or they didn't automatically marshal / unmarshal passed structs, or the interface had too many methods, making the project seem too complex to maintain and extend, proven by some that were abandoned or forked (splitting the community with it).

Related projects
----------------

- [libkv](https://github.com/docker/libkv)
    - Uses `[]byte` as value, no automatic (un-)marshalling of structs
    - No support for Redis
    - Not actively maintained anymore (3 direct commits + 1 merged PR in the last 10+ months, as of 2018-10-13)
- [valkeyrie](https://github.com/abronan/valkeyrie)
    - Fork of libkv
    - Same disadvantage: Uses `[]byte` as value, no automatic (un-)marshalling of structs
- [gokvstores](https://github.com/ulule/gokvstores)
    - Only supports Redis and local in-memory cache
    - Not actively maintained anymore (4 direct commits + 1 merged PR in the last 10+ months, as of 2018-10-13)
    - Only 13 stars (as of 2018-10-13)
- [gokv](https://github.com/gokv)
    - Requires a `json.Marshaler` / `json.Unmarshaler` as parameter, so you always need to explicitly implement their methods for your structs
    - Separate repo for each implementation, which has advantages and disadvantages
    - Single contributor
    - No releases (makes it harder to use with package managers like dep)
    - Only 2-7 stars (depending on the repository, as of 2018-10-13)
    - No support for Bolt DB / bbolt
