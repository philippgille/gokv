package combiner

// UpdateStrategy is the strategy the combiner uses when working with the configured stores when `Set()` or `Delete()` are called.
type UpdateStrategy int

const (
	// UpdateSequentialWaitAll blocks until the operation is finished for all stores.
	// Errors are collected and returned as `combiner.MultiError`.
	// This means that even if an error is returned, if the number of errors in the MultiError
	// is lower than the number of configured stores, some stores' Set/Delete operations were successful.
	UpdateSequentialWaitAll UpdateStrategy = iota + 1 // Not 0, so it's not accidentally a default value
	// UpdateParallelWaitAll is similar to UpdateSequentialWaitAll with the only difference
	// that the Set and Delete operations are forwarded to the configured stores in parallel.
	// The operation blocks until all goroutines are finished.
	// Like with UpdateSequentialWaitAll, errors are collected and returned as `combiner.MultiError`.
	// Like with UpdateSequentialWaitAll, this means that even if an error is returned,
	// if the number of errors in the MultiError is lower than the number of configured stores,
	// some stores' Set/Delete operations were successful.
	UpdateParallelWaitAll
	// UpdateSequentialWaitErrorThenContinue blocks until the operation is either successfully finished for all stores or until any single error occurs.
	// Upon error, the error is returned and all remaining stores' operations executed in a single goroutine, ignoring any errors.
	UpdateSequentialWaitErrorThenContinue
	// UpdateParallelWaitErrorThenContinue is similar to UpdateSequentialWaitErrorThenContinue with the only difference
	// that the Set and Delete operations are forwarded to the configured stores in parallel.
	// If there are no errors, the operation blocks until all goroutines are finished.
	// Upon error, the error is immediately returned, but the remaining goroutines aren't canceled.
	UpdateParallelWaitErrorThenContinue
	// UpdateSequentialWaitErrorThenSkip blocks until the operation is either successfully finished for all stores or until any single error occurs.
	// Upon error, all remaining stores' operations are skipped.
	UpdateSequentialWaitErrorThenSkip
	// UpdateParallelWaitErrorThenSkip is similar to UpdateSequentialWaitErrorThenSkip with the only difference
	// that the Set and Delete operations are forwarded to the configured stores in parallel.
	// If there are no errors, the operation blocks until all goroutines are finished.
	// Upon error, the operations in the remaining goroutines are canceled.
	UpdateParallelWaitErrorThenSkip
	// UpdateSequentialWaitNoError blocks until a store doesn't return an error.
	// This means that the combiner's Set/Delete call doesn't return an error as long as one configured store doesn't return an error.
	// Errors are ignored until the first success (no error), then the remaining stores' operations
	// are executed sequentially in a single goroutine, ignoring any errors.
	// Only if all operations lead to an error, a `combiner.MultiError` is returned, containing the list of errors.
	UpdateSequentialWaitNoError
	// UpdateSequentialWaitFirst only blocks until the first store is finished with the operation, independent of an error.
	// Upon success (no error), all remaining stores' operations are executed sequentially in a single goroutine, ignoring their results.
	// Upon error, all remaining stores' operations are skipped.
	UpdateSequentialWaitFirst
)

// GetStrategy is the strategy the combiner uses when working with the configured stores when `Get()` is called.
type GetStrategy int

const (
	// GetSequentialWaitError blocks until the operation is either successfully finished for all stores or until any single error occurs.
	// As long as no error occurs, it makes sure all fowarded Get calls either find no result
	// or find a result and they're deeply equal (via reflection).
	// Upon error, it immediately returns that error and skips any remaining stores.
	GetSequentialWaitError GetStrategy = iota + 1 // Not 0, so it's not accidentally a default value
	// GetSequentialWaitValue blocks until a store returns a value.
	// It ignores both errors and results where no value was found.
	// If no configured store returns a value, the combiner's Get operation returns `false, nil`, so "not found and no error".
	// Only if all operations lead to an error, a `combiner.MultiError` is returned, containing the list of all errors.
	GetSequentialWaitValue
	// GetSequentialWaitNoError blocks until a store doesn't return an error.
	// This means that the combiner's Get call doesn't return an error as long as one configured store doesn't return an error.
	// Errors are ignored until the first success (no error), all remaining stores' operations are skipped.
	// Only if all operations lead to an error, a `combiner.MultiError` is returned, containing the list of all errors.
	GetSequentialWaitNoError
	// GetSequentialWaitFirst only forwards the call to the first store in the stores slice, independent of an error.
	// The first store's result is returned, all other stores are always skipped.
	GetSequentialWaitFirst
)

// CloseStrategy is the strategy the combiner uses when working with the configured stores when `Close()` is called.
type CloseStrategy int

const (
	// CloseSequentialWaitAll blocks until the operation is finished for all stores.
	// Errors are collected and returned as `combiner.MultiError`.
	CloseSequentialWaitAll CloseStrategy = iota + 1 // Not 0, so it's not accidentally a default value
	// CloseParallelWaitAll is similar to CloseSequentialWaitAll with the only difference
	// that the Close operation is forwarded to the configured stores in parallel.
	// The operation blocks until all goroutines are finished.
	// If no errors occur, no error is returned.
	// Otherwise errors are collected and returned as `combiner.MultiError`.
	CloseParallelWaitAll
)
