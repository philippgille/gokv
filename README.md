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
}
```

There are detailed descriptions of the methods in the [docs](https://www.godoc.org/github.com/philippgille/gokv#Store) and in the [code](https://github.com/philippgille/gokv/blob/master/store.go). You should read them if you plan to write your own `gokv.Store` implementation or if you create a Go package with a method that takes a `gokv.Store` as parameter, so you know exactly what happens in the background.

### Implementations

[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](http://makeapullrequest.com)

- Local in-memory
    - [X] Go map (`sync.Map`)
        - Faster then a regular map when there are very few writes but lots of reads
    - [X] Go map (`map[string]byte[]` with `sync.RWMutex`)
- Embedded
    - [X] [bbolt](https://github.com/etcd-io/bbolt) (formerly known as [Bolt / Bolt DB](https://github.com/boltdb/bolt))
        - bbolt is a fork of Bolt which was maintained by CoreOS, and now by Red Hat (since CoreOS was acquired by them)
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

Design decisions
----------------

- `gokv` is primarily an abstraction for key-value stores, not caches, so there's no need for cache eviction options and timeouts.
- The package should be usable without having to write additional code, so structs should be automatically (un-)marshalled, without having to implement `Marshal()` and `Unmarshal()` first. It's still possible to implement those methods to customize the (un-)marshalling, for example to include non-exported fields, or for higher performance (because the `encoding/json` package doesn't have to use reflection).
- It should be easy to create your own store implementations, as well as to review and maintain the code of this repository, so there should be as few interface methods as possible, but still enough so that functions taking the `gokv.Store` interface as parameter can do everything that's usually required when working with a key-value store. For example, a `Watch(key string) (<-chan Notification, error)` method that sends notifications via a Go channel when the value of a given key changes is nice to have for a few use cases, but in most cases it's not required.
    - In the future we might add another interface, so that there's one for the basic operations and one for advanced uses.
- Similar projects name the structs that are implementations of the store interface according to the backing store, for example `boltdb.BoltDB`, but this leads to so called "stuttering" that's discouraged when writing idiomatic Go. That's why `gokv` uses for example `bolt.Store` and `syncmap.Store`. For easier differentiation between embedded DBs and DBs that have a client and a server component though, the first ones are called `Store` and the latter ones are called `Client`, for example `redis.Client`.

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
