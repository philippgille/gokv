gokv
====

[![GoDoc](http://www.godoc.org/github.com/philippgille/gokv?status.svg)](http://www.godoc.org/github.com/philippgille/gokv) [![Build Status](https://travis-ci.org/philippgille/gokv.svg?branch=master)](https://travis-ci.org/philippgille/gokv) [![Go Report Card](https://goreportcard.com/badge/github.com/philippgille/gokv)](https://goreportcard.com/report/github.com/philippgille/gokv) [![GitHub Releases](https://img.shields.io/github/release/philippgille/gokv.svg)](https://github.com/philippgille/gokv/releases)

Simple key-value store abstraction and implementations for Go

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
