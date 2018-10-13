Releases
========

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/) and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

vNext
-----

v0.1.0 (2018-10-14)
-------------------

Initial release with code from [philippgille/ln-paywall:78fd1dfbf10f549a22f4f30ac7f68c2a2735e989](https://github.com/philippgille/ln-paywall/tree/78fd1dfbf10f549a22f4f30ac7f68c2a2735e989) with only a few changes like a different default path and a bucket name as additional option for the Bolt DB implementation.

Features:

- Interface with `Set(string, interface{}) error` and `Get(string, interface{}) (bool, error)`
- Implementations for:
    - [bbolt](https://github.com/etcd-io/bbolt) (formerly known as Bolt / Bolt DB)
    - Go map (`sync.Map`)
    - [Redis](https://github.com/antirez/redis)
