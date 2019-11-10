package combiner

// UpdateStrategy is the strategy the combiner uses when working with the configured stores when `Set()` or `Delete()` are called.
type UpdateStrategy int

const (
	// UpdateSequentialWaitAll blocks until the operation is successfully finished for all stores.
	// It makes sure all Set and Delete operations are successful.
	// It returns early upon encountering any error.
	UpdateSequentialWaitAll UpdateStrategy = iota + 1 // Not 0, so it's not accidentally a default value
	// UpdateParallelWaitAll is similar to UpdateSequentialWaitAll with the only difference
	// that the Set and Delete operations are forwarded to the configured stores in parallel.
	// The operation blocks until all goroutines are finished, leading to the same guarantees
	// (no error returned means all operations were successful).
	// There's also the same early return in case of an error, so as soon as an error is encountered,
	// the operations in the remaining goroutines are canceled.
	UpdateParallelWaitAll
	// UpdateSequentialWaitFirst only blocks until the first store is finished with the operation, independent of an error.
	// Upon success (no error), all remaining stores' operations are executed sequentially in a single goroutine, ignoring their results.
	// Upon error, all remaining stores' operations are skipped.
	//
	// Note: It's up to the package user to either use the fastest or most reliable store as first store in the stores slice.
	UpdateSequentialWaitFirst
	// UpdateSequentialWaitSuccess blocks until the first successful operation (no error) by any store.
	// This means that a combiner operation is seen as successful even if some (but not all) stores' operations lead to errors.
	// Errors are ignored until the first success, then the remaining stores' operations are executed sequentially
	// in a single goroutine, ignoring their results.
	// Only if all operations lead to an error, a `combiner.MultiError` is returned, containing the list of errors.
	//
	// Note: It's up to the package user to either use the fastest or most reliable store as first store in the stores slice.
	UpdateSequentialWaitSuccess
)

// GetStrategy is the strategy the combiner uses when working with the configured stores when `Get()` is called.
type GetStrategy int

const (
	// GetSequentialWaitAll blocks until the operation is successfully finished for all stores.
	// It makes sure all Get calls either find no result or find a result and they're deeply equal.
	// It returns early upon encountering any error.
	GetSequentialWaitAll GetStrategy = iota + 1 // Not 0, so it's not accidentally a default value
	// GetSequentialWaitFirst only forwards the call to the first store in the stores slice, independent of an error.
	// The first store's result is returned, all other stores are always skipped.
	//
	// Note: It's up to the package user to either use the fastest or most reliable store
	// as first store in the stores slice.
	GetSequentialWaitFirst
	// GetSequentialWaitSuccess blocks until the first successful operation (no error) by any store.
	// This means that a combiner operation is seen as successful even if some (but not all) stores' operations lead to errors.
	// Errors are ignored until the first success, all remaining stores' operations are skipped.
	// Only if all operations lead to an error, a `combiner.MultiError` is returned, containing the list of errors.
	//
	// Note: It's up to the package user to either use the fastest or most reliable store as first store in the stores slice.
	GetSequentialWaitSuccess
	// GetSequentialWaitResult is similar to GetSequentialWaitSuccess with the only difference
	// that going through the store slice doesn't stop upon success (i.e. no error returned) but upon a found entry.
	// Only if all stores return no entry and no error, Get also returns no entry and no error.
	//
	// Note: It's up to the package user to either use the fastest or most reliable store as first store in the stores slice.
	GetSequentialWaitResult
)

// CloseStrategy is the strategy the combiner uses when working with the configured stores when `Close()` is called.
type CloseStrategy int

const (
	// CloseSequentialWaitAll blocks until the operation is successfully finished for all stores.
	// It returns early upon encountering any error.
	CloseSequentialWaitAll CloseStrategy = iota + 1 // Not 0, so it's not accidentally a default value
	// CloseParallelWaitAll is similar to CloseSequentialWaitAll with the only difference
	// that the Close operation is forwarded to the configured stores in parallel.
	// The operation blocks until all goroutines are finished, leading to the same guarantees
	// (no error returned means all ops successful).
	// There's also the same early return in case of an error so as soon as an error is encountered,
	// the operations in the remaining goroutines are canceled.
	CloseParallelWaitAll
)
