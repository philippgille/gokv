Choosing an implementation
==========================

If you already have a database server running and you know how to administrate it properly, you probably just want to use the appropriate implementation for that. Let's say you're already running a MySQL server. Setting up a Redis server and using the `gokv` Redis implementation will probably lead to a higher performance, but to be honest, you should always go with the service you already know well, except if you're willing to learn about a new service anyway.

Otherwise (if you *don't* already have a database server running), you might be overwhelmed with the choice of implementations.

First of all, you need to know the key differences between the store categories. Then you can look into the differences of concrete stores / implementations.

Contents
--------

1. [Categories](#categories)
2. [Implementations](#implementations)

Categories
----------

- Local in-memory
    - These are the **fastest** stores, because the data doesn't need to be transferred anywhere, not even to disk.
    - But if your application crashes, all data is lost, because it's not persisted anywhere.
    - Some of the implementations offer the possibility to limit the memory usage, leading to old data being evicted from the cache.
- Embedded
    - The data is written to disk. Easy to back up, no servers to handle.
    - Some implementations cache the data and only write to disk periodically, which leads to a high performance, but bears the risk that the data that's not persisted yet is lost when the system crashes.
    - **Well fitted for standalone client-side applications.**
    - Not fitted for web services, because you usually want to scale your services horizontally. When the DB gets filled with data by your first service instance and then you start another service instance, this second instance won't see any of the data of the first instance. Sharing the DB usually doesn't work due to open file handles and doesn't make sense anyway, because your services are probably (and should be) located on different servers with different disks.
- Distributed store
    - The database runs as a separate server.
    - The data is persisted (either immediately or periodically, depending on the implementation), so a server crash doesn't lead to data loss.
    - The servers can be run as cluster, leading to the data being even when a database server crashes.
    - Most implementations are specifically engineered to be key-value stores with very high performance.
    - **Perfectly fitted for web services**, because you can scale horizontally and each service instance can access the same data.
- Distributed cache
    - Similar to the distributed stores: Runs as separate server, can be run as cluster, engineered as key-value store with high performance
    - But without any persistence, so when a database server crashes, its data is lost.
        - (Some implementations offer optional persistence, but discourage it. If you need a distributed store with persistence, you should use one that's specifically engineered for that use-case, and not a cache.)
    - Can be useful if you only need to **cache values temporarily** and don't want a database to constantly grow.
- Cloud
    - Databases running in the cloud can be great if you need a database server (in-memory and embedded are not an option), but you can't or don't want to deal with the server administration.
    - Most DB-as-a-Service providers offer **high availability and redundancy across regions**, so even if an entire datacenter is unreachable for example due to some natural desaster or some powerlines being accidentally cut off (both happened in the past), your applications and web services still have access to all the data.
    - It might be more expensive than running your own database server, but maybe only when not taking the administrative overhead of managing your own servers into account.
    - It might be slower than running your own database server, depending on if your applications / web services run in the same cloud or not, due to network latency.
- SQL, NoSQL, NewSQL
    - Most of these implementations are probably only interesting for people who are **already running their own database servers**.

Implementations
---------------


- Local in-memory
    - Go `sync.Map`
        - Faster then a regular map when there are lots of reads and only very few writes
    - Go `map` (with `sync.RWMutex`)
    - [FreeCache](https://github.com/coocood/freecache)
        - Zero GC cache with strictly limited memory usage
        - > Note: Old entries are evicted from the cache when the cache's size limit is reached
    - [BigCache](https://github.com/allegro/bigcache)
        - Similar to FreeCache in that no GC is required even for gigabytes of data, but the memory limit is optional
        - Difference according to the BigCache creators: [BigCache vs. FreeCache](https://github.com/allegro/bigcache/blob/bff00e20c68d9f136477d62d182a7dc917bae0ca/README.md#bigcache-vs-freecache)
- Embedded
    - [bbolt](https://github.com/etcd-io/bbolt) (formerly known as [Bolt / Bolt DB](https://github.com/boltdb/bolt))
        - bbolt is a fork of Bolt which was maintained by CoreOS, and now by Red Hat (since CoreOS was acquired by them)
        - It's used for example in [etcd](https://github.com/etcd-io/etcd) as underlying persistent store
        - It uses a B+ tree, which generally means that it's very fast for read operations
    - [BadgerDB](https://github.com/dgraph-io/badger)
        - It's used for example in [Dgraph](https://github.com/dgraph-io/dgraph), a distributed graph DB
        - It uses an LSM tree, which generally means that it's very fast for write operations
    - [LevelDB / goleveldb](https://github.com/syndtr/goleveldb)
    - Local files
        - One file per key-value pair, with the key being the filename and the value being the file content
- Distributed store
    - [Redis](https://github.com/antirez/redis)
        - [The most popular distributed key-value store](https://db-engines.com/en/ranking/key-value+store)
    - [Consul](https://github.com/hashicorp/consul)
        - Probably the most popular service registry. Has a key-value store as additional feature.
        - [Official comparison with ZooKeeper, doozerd and etcd](https://github.com/hashicorp/consul/blob/df91388b7b69e1dc5bfda76f2e67b658a99324ad/website/source/intro/vs/zookeeper.html.md)
        - > Note: Consul doesn't allow values larger than 512 KB
    - [etcd](https://github.com/etcd-io/etcd)
        - It's used for example in [Kubernetes](https://github.com/kubernetes/kubernetes)
        - [Official comparison with ZooKeeper, Consul and some NewSQL databases](https://github.com/etcd-io/etcd/blob/bda28c3ce2740ef5693ca389d34c4209e431ff92/Documentation/learning/why.md#comparison-chart)
        - > Note: *By default*, the maximum request size is 1.5 MiB and the storage size limit is 2 GB. See the [documentation](https://github.com/etcd-io/etcd/blob/73028efce7d3406a19a81efd8106903eae8f4c79/Documentation/dev-guide/limit.md).
    - [TiKV](https://github.com/tikv/tikv) (⚠️Not implemented yet!)
        - Originally created as foundation of [TiDB](https://github.com/pingcap/tidb), but acts as a proper key-value store on its own and [became a project in the CNCF](https://www.cncf.io/blog/2018/08/28/cncf-to-host-tikv-in-the-sandbox/)
- Distributed cache
    - [Memcached](https://github.com/memcached/memcached)
        - > Note: Memcached is meant to be used as LRU (Least Recently Used) cache, which means items automatically *expire* and are deleted from the server after not being used for a while. See [Memcached Wiki: Forgetting is a feature](https://github.com/memcached/memcached/wiki/Overview#forgetting-is-a-feature).
    - [Hazelcast](https://github.com/hazelcast/hazelcast) (⚠️Not implemented yet!)
- Cloud
    - [Amazon DynamoDB](https://aws.amazon.com/dynamodb/)
        - > Note: The maximum value size is 400 KB. See the [documentation](https://github.com/awsdocs/amazon-dynamodb-developer-guide/blob/c420420a59040c5b3dd44a6e59f7c9e55fc922ef/doc_source/Limits.md#string).
    - [Amazon S3](https://aws.amazon.com/s3/)
        - Also works for other S3-compatible cloud services like [DigitalOcean Spaces](https://www.digitalocean.com/products/spaces/) and [Scaleway Object Storage](https://www.scaleway.com/object-storage/), as well as for self-hosted solutions like [OpenStack Swift](https://github.com/openstack/swift), [Ceph](https://github.com/ceph/ceph) and [Minio](https://github.com/minio/minio)
        - S3 is advertised as file storage (website assets, images, videos etc.), but any blob can be stored, so the `[]byte` representation of your value that `gokv` gets after automatically marshalling it can be stored as well
        - S3 is designed for high availability and redundancy, as well as low cost (compared to regular cloud databases)
        - S3 is not meant to match the performance of dedicated key-value stores
    - [Azure Cosmos DB](https://azure.microsoft.com/en-us/services/cosmos-db/)
    - [Azure Table Storage](https://azure.microsoft.com/en-us/services/storage/tables/)
        - Not as performant, scalable, flexible as Cosmos DB: [Table Storage vs. Cosmos DB Table Storage API](https://github.com/MicrosoftDocs/azure-docs/blob/58649c6910c182cba2bfc9974baed08a6fadf413/articles/cosmos-db/table-introduction.md#table-offerings)
        - But much cheaper than Cosmos DB: [Cosmos DB pricing](https://azure.microsoft.com/en-us/pricing/details/cosmos-db/) vs. [Table Storage pricing](https://azure.microsoft.com/en-us/pricing/details/storage/tables/)
        - > Note: Maximum entity size is 1 MB.
    - [Google Cloud Datastore](https://cloud.google.com/datastore/)
    - [Google Cloud Firestore](https://cloud.google.com/firestore/)
        - Currently still in beta, but might become the successor to Cloud Datastore
- SQL
    - [MySQL](https://github.com/mysql/mysql-server)
        - [The most popular open source relational database management system](https://db-engines.com/en/ranking/relational+dbms)
    - [PostgreSQL](https://github.com/postgres/postgres)
        - Seems to be seen as more advanced, robust and performant than MySQL, especially in the Go community.
- NoSQL
    - [MongoDB](https://github.com/mongodb/mongo)
        - [The most popular non-relational database](https://db-engines.com/en/ranking)
    - [Apache Cassandra](https://github.com/apache/cassandra) (⚠️Not implemented yet!)
- NewSQL
    - [CockroachDB](https://github.com/cockroachdb/cockroach) (⚠️Not implemented yet!)
        - [Official comparison with MongoDB and PostgreSQL](https://www.cockroachlabs.com/docs/stable/cockroachdb-in-comparison.html)
    - [TiDB](https://github.com/pingcap/tidb) (⚠️Not implemented yet!)
    - [Apache Ignite](https://github.com/apache/ignite) (⚠️Not implemented yet!)
