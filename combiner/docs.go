/*
Package combiner contains an implementation of the `gokv.Store` interface
which forwards its calls to multiple configured stores,
allowing you for example to store data in a fast in-memory Go map
and durable S3-compatible cloud storage at the same time with a single call to `Set()`.

Different strategies allow you to tune the behavior.
For example you can configure the combiner to block only until the first store finished its operation,
or wait for all stores to finish.
The strategies can also differ for setting and getting values.
For example you might want to make sure `Set()` is successful for all stores,
but want to work with result of the first `Get()` that finds a value.
*/
package combiner
