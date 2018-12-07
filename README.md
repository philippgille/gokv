gokv
====

[![GoDoc](http://www.godoc.org/github.com/philippgille/gokv?status.svg)](http://www.godoc.org/github.com/philippgille/gokv) [![Build Status](https://travis-ci.org/philippgille/gokv.svg?branch=master)](https://travis-ci.org/philippgille/gokv) [![Go Report Card](https://goreportcard.com/badge/github.com/philippgille/gokv)](https://goreportcard.com/report/github.com/philippgille/gokv) [![codecov](https://codecov.io/gh/philippgille/gokv/branch/master/graph/badge.svg)](https://codecov.io/gh/philippgille/gokv) [![GitHub Releases](https://img.shields.io/github/release/philippgille/gokv.svg)](https://github.com/philippgille/gokv/releases)

Simple key-value store abstraction and implementations for Go

Contents
--------

1. [Features](#features)
    1. [Simple interface](#simple-interface)
    2. [Implementations](#implementations)
    3. [Value types](#value-types)
    4. [Marshal formats](#marshal-formats)
2. [Usage](#usage)
3. [Project status](#project-status)
4. [Motivation](#motivation)
5. [Design decisions](#design-decisions)
6. [Related projects](#related-projects)

Features
--------

### Simple interface

> Note: The interface is not final yet! See [Project status](#project-status) for details.

```go
type Store interface {
    Set(k string, v interface{}) error
    Get(k string, v interface{}) (found bool, err error)
    Delete(k string) error
    Close() error
}
```

There are detailed descriptions of the methods in the [docs](https://www.godoc.org/github.com/philippgille/gokv#Store) and in the [code](https://github.com/philippgille/gokv/blob/master/store.go). You should read them if you plan to write your own `gokv.Store` implementation or if you create a Go package with a method that takes a `gokv.Store` as parameter, so you know exactly what happens in the background.

### Implementations

Some of the following databases aren't specifically engineered for storing key-value pairs, but if someone's running them already for other purposes and doesn't want to set up one of the proper key-value stores due to administrative overhead etc., they can of course be used as well. In those cases let's focus on a few of the most popular though. This mostly goes for the SQL, NoSQL and NewSQL categories.

Feel free to suggest more stores by creating an [issue](https://github.com/philippgille/gokv/issues) or even add an actual implementation - [![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](http://makeapullrequest.com).

- Local in-memory
    - [X] Go map (`sync.Map`)
        - Faster then a regular map when there are lots of reads and only very few writes
    - [X] Go map (`map[string]byte[]` with `sync.RWMutex`)
- Embedded
    - [X] [bbolt](https://github.com/etcd-io/bbolt) (formerly known as [Bolt / Bolt DB](https://github.com/boltdb/bolt))
        - bbolt is a fork of Bolt which was maintained by CoreOS, and now by Red Hat (since CoreOS was acquired by them)
        - It's used for example in [etcd](https://github.com/etcd-io/etcd) as underlying persistent store
        - It uses a B+ tree, which generally means that it's very fast for read operations
    - [X] [BadgerDB](https://github.com/dgraph-io/badger)
        - It's used for example in [Dgraph](https://github.com/dgraph-io/dgraph), a distributed graph DB
        - It uses an LSM tree, which generally means that it's very fast for write operations
    - [ ] [LevelDB / goleveldb](https://github.com/syndtr/goleveldb)
- Distributed store
    - [X] [Redis](https://github.com/antirez/redis)
        - [The most popular distributed key-value store](https://db-engines.com/en/ranking/key-value+store)
    - [X] [Consul](https://github.com/hashicorp/consul)
        - Probably the most popular service registry. Has a key-value store as additional feature.
        - [Official comparison with ZooKeeper, doozerd and etcd](https://github.com/hashicorp/consul/blob/df91388b7b69e1dc5bfda76f2e67b658a99324ad/website/source/intro/vs/zookeeper.html.md)
        - > Note: Consul doesn't allow values larger than 512 KB
    - [X] [etcd](https://github.com/etcd-io/etcd)
        - It's used for example in [Kubernetes](https://github.com/kubernetes/kubernetes)
        - [Official comparison with ZooKeeper, Consul and some NewSQL databases](https://github.com/etcd-io/etcd/blob/bda28c3ce2740ef5693ca389d34c4209e431ff92/Documentation/learning/why.md#comparison-chart)
        - > Note: *By default*, the maximum request size is 1.5 MiB and the storage size limit is 2 GB. See the [documentation](https://github.com/etcd-io/etcd/blob/73028efce7d3406a19a81efd8106903eae8f4c79/Documentation/dev-guide/limit.md).
    - [ ] [TiKV](https://github.com/tikv/tikv)
        - Originally created as foundation of [TiDB](https://github.com/pingcap/tidb), but acts as a proper key-value store on its own and [became a project in the CNCF](https://www.cncf.io/blog/2018/08/28/cncf-to-host-tikv-in-the-sandbox/)
- Distributed cache (no presistence *by default*)
    - [X] [Memcached](https://github.com/memcached/memcached)
        - > Note: Memcached is meant to be used as LRU (Least Recently Used) cache, which means items automatically *expire* and are deleted from the server after not being used for a while. See [Memcached Wiki: Forgetting is a feature](https://github.com/memcached/memcached/wiki/Overview#forgetting-is-a-feature).
    - [ ] [Hazelcast](https://github.com/hazelcast/hazelcast)
- Cloud
    - [X] [Amazon DynamoDB](https://aws.amazon.com/dynamodb/)
        - > Note: The maximum value size is 400 KB. See the [documentation](https://github.com/awsdocs/amazon-dynamodb-developer-guide/blob/c420420a59040c5b3dd44a6e59f7c9e55fc922ef/doc_source/Limits.md#string).
    - [ ] [Azure Cosmos DB](https://azure.microsoft.com/en-us/services/cosmos-db/)
    - [X] [Azure Table Storage](https://azure.microsoft.com/en-us/services/storage/tables/)
        - Not as performant, scalable, flexible as Cosmos DB: [Table Storage vs. Cosmos DB Table Storage API](https://github.com/MicrosoftDocs/azure-docs/blob/58649c6910c182cba2bfc9974baed08a6fadf413/articles/cosmos-db/table-introduction.md#table-offerings)
        - But much cheaper than Cosmos DB: [Cosmos DB pricing](https://azure.microsoft.com/en-us/pricing/details/cosmos-db/) vs. [Table Storage pricing](https://azure.microsoft.com/en-us/pricing/details/storage/tables/)
        - > Note: Maximum entity size is 1 MB.
    - [ ] [Google Cloud Datastore](https://cloud.google.com/datastore/)
- SQL
    - [X] [MySQL](https://github.com/mysql/mysql-server)
        - [The most popular open source relational database management system](https://db-engines.com/en/ranking/relational+dbms)
    - [ ] [PostgreSQL](https://github.com/postgres/postgres)
- NoSQL
    - [X] [MongoDB](https://github.com/mongodb/mongo)
        - [The most popular non-relational database](https://db-engines.com/en/ranking)
    - [ ] [Apache Cassandra](https://github.com/apache/cassandra)
- NewSQL
    - [ ] [CockroachDB](https://github.com/cockroachdb/cockroach)
        - [Official comparison with MongoDB and PostgreSQL](https://www.cockroachlabs.com/docs/stable/cockroachdb-in-comparison.html)
    - [ ] [TiDB](https://github.com/pingcap/tidb)
    - [ ] [Apache Ignite](https://github.com/apache/ignite)

### Value types

Most Go packages for key-value stores just accept a `[]byte` as value, which requires developers for example to marshal (and later unmarshal) their structs. `gokv` is meant to be simple and make developers' lifes easier, so it accepts any type (with using `interface{}` as parameter), including structs, and automatically (un-)marshals the value.

The kind of (un-)marshalling is left to the implementation. All implementations in this repository currently support JSON and [gob](https://blog.golang.org/gobs-of-data) by using `encoding/json` and `encoding/gob`. See [Marshal formats](#marshal-formats) for details.

For unexported struct fields to be (un-)marshalled to/from JSON/gob, the respective custom (un-)marshalling methods need to be implemented as methods of the struct (e.g. `MarshalJSON() ([]byte, error)` for custom marshalling into JSON). See [Marshaler](https://godoc.org/encoding/json#Marshaler) and [Unmarshaler](https://godoc.org/encoding/json#Unmarshaler) for JSON, and [GobEncoder](https://godoc.org/encoding/gob#GobEncoder) and [GobDecoder](https://godoc.org/encoding/gob#GobDecoder) for gob.

To improve performance you can also implement the custom (un-)marshalling methods so that no reflection is used by the `encoding/json` / `encoding/gob` packages. This is not a disadvantage of using a generic key-value store package, it's the same as if you would use a concrete key-value store package which only accepts `[]byte`, requiring you to (un-)marshal your structs.

### Marshal formats

- [X] JSON
- [X] [gob](https://blog.golang.org/gobs-of-data)

The stores marshal and unmarshal the values when storing / retrieving them. The default format is JSON, but all `gokv.Store` implementations in this repository also support [gob](https://blog.golang.org/gobs-of-data) as alternative, configurable via their `Options`.

The marshal format is up to the implementations though, so package creators using the `gokv.Store` interface as parameter of a function should not make any assumptions about this. If they require any specific format they should inform the package user about this in the GoDoc of the function taking the store interface as parameter.

Differences:

- Depending on the struct, one of the formats might be faster
- Depending on the struct, one of the formats might lead to a lower storage size
- Depending on the use case, the custom (un-)marshal methods of one of the formats might be easier to implement
    - JSON: [`MarshalJSON() ([]byte, error)`](https://godoc.org/encoding/json#Marshaler) and [`UnmarshalJSON([]byte) error`](https://godoc.org/encoding/json#Unmarshaler)
    - gob: [`GobEncode() ([]byte, error)`](https://godoc.org/encoding/gob#GobEncoder) and [`GobDecode([]byte) error`](https://godoc.org/encoding/gob#GobDecoder)

Usage
-----

Every implementation has its own `Options` struct, but all implementations have a `NewStore()` / `NewClient()` function that returns an object of a sctruct that implements the `gokv.Store` interface. Let's take the implementation for Redis as example, which is the most popular distributed key-value store.

```go
package main

import (
    "fmt"

    "github.com/philippgille/gokv"
    "github.com/philippgille/gokv/redis"
)

type foo struct {
    Bar string
}

func main() {
    options := redis.DefaultOptions // Address: "localhost:6379", Password: "", DB: 0

    // Create client
    client, err := redis.NewClient(options)
    if err != nil {
        panic(err)
    }

    // Store, retrieve, print and delete a value
    interactWithStore(client)

    // Close client
    err = client.Close()
    if err != nil {
        panic(err)
    }
}

// interactWithStore stores, retrieves, prints and deletes a value.
// It's completely independent of the store implementation.
func interactWithStore(store gokv.Store) {
    // Store value
    val := foo{
        Bar: "baz",
    }
    err := store.Set("foo123", val)
    if err != nil {
        panic(err)
    }

    // Retrieve value
    retrievedVal := new(foo)
    found, err := store.Get("foo123", retrievedVal)
    if err != nil {
        panic(err)
    }
    if !found {
        panic("Value not found")
    }

    fmt.Printf("foo: %+v", *retrievedVal) // Prints `foo: {Bar:baz}`

    // Delete value
    err = store.Delete("foo123")
    if err != nil {
        panic(err)
    }
}
```

As described in the comments, that code does the following:

1. Create a client for Redis
2. Call `interactWithStore()`, which requires a `gokv.Store` as parameter. This method then:
    1. Stores an object of type `foo` in the Redis server running on `localhost:6379` with the key `foo123`
    2. Retrieves the value for the key `foo123`
        - The check if the value was found isn't needed in this example but is included for demonstration purposes
    3. Prints the value. It prints `foo: {Bar:baz}`, which is exactly what was stored before.
    4. Deletes the value

Now let's say you don't want to use Redis but Consul instead. You just have to make three simple changes:

1. Replace the import of `"github.com/philippgille/gokv/redis"` by `"github.com/philippgille/gokv/consul"`
2. Replace `redis.DefaultOptions` by `consul.DefaultOptions`
3. Replace `redis.NewClient(options)` by `consul.NewClient(options)`

Everything else works the same way. `interactWithStore()` is completely unaffected.

Project status
--------------

> Note: `gokv`'s API is not stable yet and is under active development. Upcoming releases are likely to contain breaking changes as long as the version is `v0.x.y`. You should use vendoring to prevent bad surprises. This project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html) and all notable changes to this project are documented in [RELEASES.md](https://github.com/philippgille/gokv/blob/master/RELEASES.md).

Planned interface methods until `v1.0.0`:

- `List(interface{}) error` / `GetAll(interface{}) error` or similar

The interface might even change until `v1.0.0`. For example one consideration is to change `Get(string, interface{}) (bool, error)` to `Get(string, interface{}) error` (no boolean return value anymore), with the `error` being something like `gokv.ErrNotFound // "Key-value pair not found"` to fulfill the additional role of indicating that the key-value pair wasn't found. But at the moment we prefer the current method signature.

Also, more interfaces might be added. For example so that there's a `SimpleStore` and an `AdvancedStore`, with the first one containing only the basic methods and the latter one with advanced features such as key-value pair lifetimes (deletion of key-value pairs after a given time), notification of value changes via Go channels etc. But currently the focus is simplicity, see [Design decisions](#design-decisions).

Motivation
----------

When creating a package you want the package to be usable by as many developers as possible. Let's look at a specific example: You want to create a paywall middleware for the Gin web framework. You need some database to store state. You can't use a Go map, because its data is not persisted across web service restarts. You can't use an embedded DB like bbolt, BadgerDB or SQLite, because that would restrict the web service to one instance, but nowadays every web service is designed with high horizontal scalability in mind. If you use Redis, MongoDB or PostgreSQL though, you would force the package user (the developer who creates the actual web service with Gin and your middleware) to run and administrate the server, even if she might never have used it before and doesn't know how to configure them for high performance and security.

Any decision for a specific database would limit the package's usability.

One solution would be a custom interface where you would leave the implementation to the package user. But that would require the developer to dive into the details of the Go package of the chosen key-value store. And if the developer wants to switch the store, or maybe use one for local testing and another for production, she would need to write *multiple* implementations.

`gokv` is the solution for these problems. Package *creators* use the `gokv.Store` interface as parameter and can call its methods within their code, leaving the decision which actual store to use to the package user. Package *users* pick one of the implementations, for example `github.com/philippgille/gokv/redis` for Redis and pass the `redis.Client` created by `redis.NewClient(...)` as parameter. Package users can also develop their own implementations if they need to.

`gokv` doesn't just have to be used to satisfy some `gokv.Store` parameter. It can of course also be used by application / web service developers who just don't want to dive into the sometimes complicated usage of some key-value store packages.

Initially it was developed as `storage` package within the project [ln-paywall](https://github.com/philippgille/ln-paywall) to provide the users of ln-paywall with multiple storage options, but at some point it made sense to turn it into a repository of its own.

Before doing so I examined existing Go packages with a similar purpose (see [Related projects](#related-projects)), but none of them fit my needs. They either had too few implementations, or they didn't automatically marshal / unmarshal passed structs, or the interface had too many methods, making the project seem too complex to maintain and extend, proven by some that were abandoned or forked (splitting the community with it).

Design decisions
----------------

- `gokv` is primarily an abstraction for **key-value stores**, not caches, so there's no need for cache eviction and timeouts.
    - It's still possible to have cache eviction. In some cases you can configure it on the server, or in case of Memcached it's even the default. Or you can have an implementation-specific `Option` that configures the key-value store client to set a timeout on some key-value pair when storing it in the server. But this should be implementation-specific and not be part of the interface methods, which would require *every* implementation to support cache eviction.
- The package should be usable without having to write additional code, so structs should be (un-)marshalled automatically, without having to implement `MarshalJSON()` / `GobEncode()` and `UnmarshalJSON()` / `GobDecode()` first. It's still possible to implement these methods to customize the (un-)marshalling, for example to include unexported fields, or for higher performance (because the `encoding/json` / `encoding/gob` package doesn't have to use reflection).
- It should be easy to create your own store implementations, as well as to review and maintain the code of this repository, so there should be as few interface methods as possible, but still enough so that functions taking the `gokv.Store` interface as parameter can do everything that's usually required when working with a key-value store. For example, a boolean return value for the `Delete` method that indicates whether a value was actually deleted (because it was previously present) can be useful, but isn't a must-have, and also it would require some `Store` implementations to implement the check by themselves (because the existing libraries don't support it), which would unnecessarily decrease performance for those who don't need it. Or as another example, a `Watch(key string) (<-chan Notification, error)` method that sends notifications via a Go channel when the value of a given key changes is nice to have for a few use cases, but in most cases it's not required.
    - > Note: In the future we might add another interface, so that there's one for the basic operations and one for advanced uses.
- Similar projects name the structs that are implementations of the store interface according to the backing store, for example `boltdb.BoltDB`, but this leads to so called "stuttering" that's discouraged when writing idiomatic Go. That's why `gokv` uses for example `bbolt.Store` and `syncmap.Store`. For easier differentiation between embedded DBs and DBs that have a client and a server component though, the first ones are called `Store` and the latter ones are called `Client`, for example `redis.Client`.
- All errors are implementation-specific. We could introduce a `gokv.StoreError` type and define some constants like a `SetError` or something more specific like a `TimeoutError`, but non-specific errors don't help the package user, and specific errors would make it very hard to create and especially maintain a `gokv.Store` implementation. You would need to know exactly in which cases the package (that the implementation uses) returns errors, what the errors mean (to "translate" them) and keep up with changes and additions of errors in the package. So instead, errors are just forwarded. For example, if you use the `dynamodb` package, the returned errors will be errors from the `"github.com/aws/aws-sdk-go` package.

Related projects
----------------

- [libkv](https://github.com/docker/libkv)
    - Uses `[]byte` as value, no automatic (un-)marshalling of structs
    - No support for Redis, BadgerDB, Go map, MongoDB, AWS DynamoDB, Memcached, MySQL, ...
    - Not actively maintained anymore (3 direct commits + 1 merged PR in the last 10+ months, as of 2018-10-13)
- [valkeyrie](https://github.com/abronan/valkeyrie)
    - Fork of libkv
    - Same disadvantage: Uses `[]byte` as value, no automatic (un-)marshalling of structs
    - No support for BadgerDB, Go map, MongoDB, AWS DynamoDB, Memcached, MySQL, ...
- [gokvstores](https://github.com/ulule/gokvstores)
    - Only supports Redis and local in-memory cache
    - Not actively maintained anymore (4 direct commits + 1 merged PR in the last 10+ months, as of 2018-10-13)
    - 13 stars (as of 2018-10-13)
- [gokv](https://github.com/gokv)
    - Requires a `json.Marshaler` / `json.Unmarshaler` as parameter, so you always need to explicitly implement their methods for your structs, and also you can't use gob or other formats for (un-)marshaling.
    - No support for Consul, etcd, bbolt / Bolt, BadgerDB, MongoDB, AWS DynamoDB, Memcached, MySQL, ...
    - Separate repo for each implementation, which has advantages and disadvantages
    - No releases (makes it harder to use with package managers like dep)
    - 2-7 stars (depending on the repository, as of 2018-10-13)

Others:

- [gladkikhartem/gokv](https://github.com/gladkikhartem/gokv): No `Delete()` method, no Redis, embedded DBs etc., no Git tags / releases, no stars (as of 2018-11-28)
- [bradberger/gokv](https://github.com/bradberger/gokv): Not maintained (no commits in the last 22 months), no Redis, Consul etc., no Git tags / releases, 1 star (as of 2018-11-28)
- [ppacher/gokv](https://github.com/ppacher/gokv): Not maintained (no commits in the last 22 months), no Redis, embedded DBs etc., no automatic (un-)marshalling, 1 star (as of 2018-11-28)
    - Nice CLI though!
- [kapitan-k/gokvstore](https://github.com/kapitan-k/gokvstore): Not actively maintained (no commits in the last 10+ months), RocksDB only, requires cgo, no automatic (un-)marshalling, no Git tags/ releases, 1 star (as of 2018-11-28)
