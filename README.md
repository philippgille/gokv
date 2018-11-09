gokv
====

[![GoDoc](http://www.godoc.org/github.com/philippgille/gokv?status.svg)](http://www.godoc.org/github.com/philippgille/gokv) [![Build Status](https://travis-ci.org/philippgille/gokv.svg?branch=master)](https://travis-ci.org/philippgille/gokv) [![Go Report Card](https://goreportcard.com/badge/github.com/philippgille/gokv)](https://goreportcard.com/report/github.com/philippgille/gokv) [![codecov](https://codecov.io/gh/philippgille/gokv/branch/master/graph/badge.svg)](https://codecov.io/gh/philippgille/gokv) [![GitHub Releases](https://img.shields.io/github/release/philippgille/gokv.svg)](https://github.com/philippgille/gokv/releases)

Simple key-value store abstraction and implementations for Go

Features
--------

### Simple interface

> Note: The interface is not final yet! See [Project status](#project-status) for details.

```go
type Store interface {
	Set(string, interface{}) error
	Get(string, interface{}) (bool, error)
	Delete(string) error
}
```

There are detailed descriptions of the methods in the [docs](https://www.godoc.org/github.com/philippgille/gokv#Store) and in the [code](https://github.com/philippgille/gokv/blob/master/store.go). You should read them if you plan to write your own `gokv.Store` implementation or if you create a Go package with a method that takes a `gokv.Store` as parameter, so you know exactly what happens in the background.

### Value types

Most Go packages for key-value stores just accept a `[]byte` as value, which requires developers for example to marshal (and later unmarshal) their structs. `gokv` is meant to be simple and make developers' lifes easier, so it accepts any type (with using `interface{}` as parameter), including structs, and automatically (un-)marshals the value.

The kind of (un-)marshalling is left to the implementation. All implementations in `github.com/philippgille/gokv` currently support JSON by using `encoding/json`, but `encoding/gob` will be added as alternative in the future, which will have some advantages and some tradeoffs.

For unexported struct fields to be (un-)marshalled to/from JSON, `UnmarshalJSON(b []byte) error` and `MarshalJSON() ([]byte, error)` need to be implemented as methods of the struct.

To improve performance you can also implement the `UnmarshalJSON()` and `MarshalJSON()` methods so that no reflection is used by the `encoding/json` package. This is the same as if you would use a key-value store package which only accepts `[]byte`, requiring you to (un-)marshal your structs.

### Implementations

Some databases aren't specifically engineered for storing key-value pairs, but if someone's running them already for other purposes and doesn't want to set up one of the proper key-value stores due to administrative overhead etc., they can of course be used as well. In those cases let's focus on a few of the most popular though.

[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](http://makeapullrequest.com)

- Local in-memory
    - [X] Go map (`sync.Map`)
        - Faster then a regular map when there are very few writes but lots of reads
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
    - [X] [Consul](https://github.com/hashicorp/consul)
        - > Note: Consul doesn't allow values larger than 512 KB
    - [ ] [etcd](https://github.com/etcd-io/etcd)
        - Not advertised as general key-value store, but can be used as one
    - [ ] [TiKV](https://github.com/tikv/tikv)
        - Originally created to complement [TiDB](https://github.com/pingcap/tidb), but recently [became a project in the CNCF](https://www.cncf.io/blog/2018/08/28/cncf-to-host-tikv-in-the-sandbox/)
    - [ ] [LedisDB](https://github.com/siddontang/ledisdb)
        - Similar to Redis, with several backing stores
- Distributed cache (no presistence *by default*)
    - [ ] [Memcached](https://github.com/memcached/memcached)
    - [ ] [Hazelcast](https://github.com/hazelcast/hazelcast)
- Cloud
    - [ ] [Amazon DynamoDB](https://aws.amazon.com/dynamodb/)
    - [ ] [Amazon SimpleDB](https://aws.amazon.com/simpledb/)
    - [ ] [Azure Cosmos DB](https://azure.microsoft.com/en-us/services/cosmos-db/)
    - [ ] [Azure Table Storage](https://azure.microsoft.com/en-us/services/storage/tables/)
    - [ ] [Google Cloud Datastore](https://cloud.google.com/datastore/)
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

- `List(interface{}) error` / `GetAll(interface{}) error` or similar
- `Close() error` or similar

Motivation
----------

When creating a package you want the package to be usable by as many developers as possible. Let's look at a specific example: You want to create a paywall middleware for the Gin web framework. You need some database to store state. You can't use a Go map, because its data is not persisted across web service restarts. You can't use an embedded DB like bbolt, BadgerDB or SQLite, because that would restrict the web service to one instance, but nowadays every web service is designed with high horizontal scalability in mind. If you use Redis or Consul though, you would force the package user (the developer who creates the actual web service with Gin and your middleware) to run and administrate a Redis or Consul server, even if she might never have used it before and doesn't know how to manage their persistence and security models.

One solution would be a custom interface where you would leave the implementation to the package user. But that would require the developer to dive into the details of the Go package of the chosen key-value store. And if the developer wants to switch the store, or maybe use one for local testing and another for production, she would need to write *multiple* implementations.

`gokv` is the solution for these problems. Package *creators* use the `gokv.Store` interface as parameter and can call its methods within their code, leaving the decision which actual store to use to the package user. Package *users* pick one of the implementations, for example `github.com/philippgille/gokv/redis` for Redis and pass the `redis.Client` created by `redis.NewClient(...)` as parameter. Package users can also develop their own implementations if they need to.

`gokv` can of course also be used by application / web service developers who just don't want to dive into the sometimes complicated usage of some key-value store packages.

Initially it was developed as `storage` package within the project [ln-paywall](https://github.com/philippgille/ln-paywall) to provide the users of ln-paywall with multiple storage options, but at some point it made sense to turn it into a repository of its own.

Before doing so I examined existing Go packages with a similar purpose (see [Related projects](#related-projects), but none of them fit my needs. They either had too few implementations, or they didn't automatically marshal / unmarshal passed structs, or the interface had too many methods, making the project seem too complex to maintain and extend, proven by some that were abandoned or forked (splitting the community with it).

Design decisions
----------------

- `gokv` is primarily an abstraction for key-value stores, not caches, so there's no need for cache eviction options and timeouts.
- The package should be usable without having to write additional code, so structs should be automatically (un-)marshalled, without having to implement `MarshalJSON()` and `UnmarshalJSON()` first. It's still possible to implement these methods to customize the (un-)marshalling, for example to include unexported fields, or for higher performance (because the `encoding/json` package doesn't have to use reflection).
- It should be easy to create your own store implementations, as well as to review and maintain the code of this repository, so there should be as few interface methods as possible, but still enough so that functions taking the `gokv.Store` interface as parameter can do everything that's usually required when working with a key-value store. For example, a boolean return value for the `Delete` method that indicates whether a value was actually deleted (because it was previously present) can be useful, but isn't a must-have, and also it would require some `Store` implementations to implement the check by themselves (because the existing libraries don't support it), which would unnecessarily decrease performance for those who don't need it. Or as another example, a `Watch(key string) (<-chan Notification, error)` method that sends notifications via a Go channel when the value of a given key changes is nice to have for a few use cases, but in most cases it's not required.
    - In the future we might add another interface, so that there's one for the basic operations and one for advanced uses.
- Similar projects name the structs that are implementations of the store interface according to the backing store, for example `boltdb.BoltDB`, but this leads to so called "stuttering" that's discouraged when writing idiomatic Go. That's why `gokv` uses for example `bolt.Store` and `syncmap.Store`. For easier differentiation between embedded DBs and DBs that have a client and a server component though, the first ones are called `Store` and the latter ones are called `Client`, for example `redis.Client`.

Related projects
----------------

- [libkv](https://github.com/docker/libkv)
    - Uses `[]byte` as value, no automatic (un-)marshalling of structs
    - No support for Redis, BadgerDB, Go map, ...
    - Not actively maintained anymore (3 direct commits + 1 merged PR in the last 10+ months, as of 2018-10-13)
- [valkeyrie](https://github.com/abronan/valkeyrie)
    - Fork of libkv
    - Same disadvantage: Uses `[]byte` as value, no automatic (un-)marshalling of structs
    - No support for BadgerDB, Go map, ...
- [gokvstores](https://github.com/ulule/gokvstores)
    - Only supports Redis and local in-memory cache
    - Not actively maintained anymore (4 direct commits + 1 merged PR in the last 10+ months, as of 2018-10-13)
    - Only 13 stars (as of 2018-10-13)
- [gokv](https://github.com/gokv)
    - Requires a `json.Marshaler` / `json.Unmarshaler` as parameter, so you always need to explicitly implement their methods for your structs, and also you can't use gob for (un-)marshaling.
    - Separate repo for each implementation, which has advantages and disadvantages
    - Single contributor
    - No releases (makes it harder to use with package managers like dep)
    - Only 2-7 stars (depending on the repository, as of 2018-10-13)
    - No support for Consul, Bolt DB / bbolt, BadgerDB, ...
