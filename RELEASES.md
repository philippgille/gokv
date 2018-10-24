Releases
========

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/) and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

vNext
-----

- Added: Package `gomap` - A `gokv.Store` implementation for a plain Go map with a `sync.RWMutex` for concurrent access (issue [#11](https://github.com/philippgille/gokv/issues/11))
- Improved: Every `gokv.Store` implementation resides in its own package now, so when downloading the package of an implementation, for example with `go get github.com/philippgille/gokv/redis`, only the actually required dependencies are downloaded and compiled, making the process much faster. This is especially useful for example when creating Docker images, where in many cases (depending on the `Dockerfile`) the download and compilation are repeated for *each build*. (Issue [#2](https://github.com/philippgille/gokv/issues/2))
- Improved: The performance of BoltClient should be higher, because unnecessary manual locking was removed. (Issue [#1](https://github.com/philippgille/gokv/issues/1))

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
