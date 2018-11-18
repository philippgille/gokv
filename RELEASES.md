Releases
========

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/) and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

vNext
-----

- Added: Package `dynamodb` - A `gokv.Store` implementation for [Amazon DynamoDB](https://aws.amazon.com/dynamodb/) (issue [#28](https://github.com/philippgille/gokv/issues/28))
- Fixed spelling in error message when using the etcd implementation and the etcd server is unreachable

v0.3.0 (2018-11-17)
-------------------

- Added: Method `Delete(string) error` (issue [#8](https://github.com/philippgille/gokv/issues/8))
- Added: All `gokv.Store` implementations in this package now also support [gob](https://blog.golang.org/gobs-of-data) as marshal format as alternative to JSON (issue [#22](https://github.com/philippgille/gokv/issues/22))
    - Part of this addition are a new field in the existing `Options` structs, called `MarshalFormat`, as well as the related `MarshalFormat` enum (custom type + related `const` values) in each implementation package
- Added: Package `badgerdb` - A `gokv.Store` implementation for [BadgerDB](https://github.com/dgraph-io/badger) (issue [#16](https://github.com/philippgille/gokv/issues/16))
- Added: Package `consul` - A `gokv.Store` implementation for [Consul](https://github.com/hashicorp/consul) (issue [#18](https://github.com/philippgille/gokv/issues/18))
- Added: Package `etcd` - A `gokv.Store` implementation for [etcd](https://github.com/etcd-io/etcd) (issue [#24](https://github.com/philippgille/gokv/issues/24))

### Breaking changes

- Changed: The `NewStore()` function in `gomap` and `syncmap` now has an `Option` parameter. Required for issue [#22](https://github.com/philippgille/gokv/issues/22).
- Changed: Passing an empty string as key to `Set()`, `Get()` or `Delete()` now results in an error
- Changed: Passing `nil` as value parameter to `Set()` or as pointer to `Get()` now results in an error. This change leads to a consistent behaviour across the different marshal formats (otherwise for example `encoding/json` marshals `nil` to `null` while `encoding/gob` returns an error).

v0.2.0 (2018-11-05)
-------------------

- Added: Package `gomap` - A `gokv.Store` implementation for a plain Go map with a `sync.RWMutex` for concurrent access (issue [#11](https://github.com/philippgille/gokv/issues/11))
- Improved: Every `gokv.Store` implementation resides in its own package now, so when downloading the package of an implementation, for example with `go get github.com/philippgille/gokv/redis`, only the actually required dependencies are downloaded and compiled, making the process much faster. This is especially useful for example when creating Docker images, where in many cases (depending on the `Dockerfile`) the download and compilation are repeated for *each build*. (Issue [#2](https://github.com/philippgille/gokv/issues/2))
- Improved: The performance of `bolt.Store` should be higher, because unnecessary manual locking was removed. (Issue [#1](https://github.com/philippgille/gokv/issues/1))
- Fixed: The `gokv.Store` implementation for bbolt / Bolt DB used data from within a Bolt transaction outside of it, without copying the value, which can lead to errors (see [here](https://github.com/etcd-io/bbolt/blob/76a4670663d125b6b89d47ea3cc659a282d87c28/doc.go#L38)) (issue [#13](https://github.com/philippgille/gokv/issues/13))

### Breaking changes

- All `gokv.Store` implementations were moved into their own packages and the structs that implement the interface were renamed to avoid unidiomatic "stuttering".

v0.1.0 (2018-10-14)
-------------------

Initial release with code from [philippgille/ln-paywall:78fd1dfbf10f549a22f4f30ac7f68c2a2735e989](https://github.com/philippgille/ln-paywall/tree/78fd1dfbf10f549a22f4f30ac7f68c2a2735e989) with only a few changes like a different default path and a bucket name as additional option for the Bolt DB implementation.

Features:

- Interface with `Set(string, interface{}) error` and `Get(string, interface{}) (bool, error)`
- Implementations for:
    - [bbolt](https://github.com/etcd-io/bbolt) (formerly known as Bolt / Bolt DB)
    - Go map (`sync.Map`)
    - [Redis](https://github.com/antirez/redis)
