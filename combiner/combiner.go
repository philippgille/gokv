package combiner

import (
	"context"
	"errors"
	"reflect"
	"runtime"

	"github.com/philippgille/gokv"
	"github.com/philippgille/gokv/util"
)

// Ensure that we satisfy interface.
var _ gokv.Store = Combiner{}

// Strategy is the strategy with which the combiner should work with the configured stores.
type Strategy int

const (
	// SequentialWaitAll is the most reliable strategy.
	// It blocks until the operation is successfully finished for all stores.
	// It makes sure all Set and Delete operations are successful,
	// and makes sure all Get calls find a result and they're deeply equal.
	// It returns early upon encountering any error.
	SequentialWaitAll Strategy = iota // 0, so it's the default value for a Strategy
	// ParallelWaitAll is similar to SequentialWaitAll with the only difference
	// that the Set/Get/Delete/Close operations are forwarded to the configured stores
	// in parallel. The operation still blocks until all goroutines are finished,
	// leading to the same guarantees (no error returned means all ops successful, all results deeply equal).
	// There's also the same early return in case of an error.
	// So as soon as an error is encountered, the operations in the remaining goroutines are canceled.
	ParallelWaitAll
	// SequentialWaitFirst only blocks until the first store is finished with the operation,
	// independent of a success.
	//
	// For Set and Delete: Upon success, all remaining stores' operations
	// are executed sequentially in a single goroutine, ignoring their results.
	// For Get: All remaining stores' operations are skipped.
	// So essentially Get is always only called on the first store.
	//
	// Note: It's up to the package user to either use the fastest or most reliable store
	// as first store in the stores slice.
	SequentialWaitFirst
	// SequentialWaitSuccess blocks until the first successful operation by any store.
	// This means that a combiner operation is seen as successful even if some (but not all) stores' operation is unsuccessful.
	//
	// For Set and Delete: Failures are ignored until the first success,
	// then the remaining stores' operations are executed sequentially in a single goroutine,
	// ignoring their results.
	// If all operations fail, a `combiner.MultiError` is returned, containing the list of errors.
	// For Get: Failures are ignored until the first success, all remaining stores' operations are skipped.
	// If all operations fail, a `combiner.MultiError` is returned, containing the list of errors.
	// Note: When a value is not found during Get, this is not seen as failure. Only errors are failures.
	//
	// Note: It's up to the package user to either use the fastest or most reliable store
	// as first store in the stores slice.
	SequentialWaitSuccess
	// SequentialWaitResult is similar to SequentialWaitSuccess with the only difference
	// that Get doesn't stop upon success (i.e. no error returned) but upon a found entry.
	// Only if all stores return no entry and no error, Get also returns no entry and no error.
	//
	// Note: It's up to the package user to either use the fastest or most reliable store
	// as first store in the stores slice.
	SequentialWaitResult
)

// Combiner is a `gokv.Store` implementation that forwards its
// calls to other `gokv.Store`s with configurable strategies.
// At least two stores must be set.
type Combiner struct {
	stores         []gokv.Store
	setStrategy    Strategy
	getStrategy    Strategy
	deleteStrategy Strategy
	closeStrategy  Strategy
}

// Set stores the given value for the given key.
// Values are automatically marshalled to JSON or gob (depending on the configuration).
// The key must not be "" and the value must not be nil.
// The call is forwarded to the configured stores according to the configured strategy.
// Returned errors are of type `combiner.MultiError`
func (s Combiner) Set(k string, v interface{}) error {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return newMultiError(err)
	}

	switch s.setStrategy {
	case SequentialWaitAll:
		for _, store := range s.stores {
			err := store.Set(k, v)
			if err != nil {
				return newMultiError(err)
			}
		}
	case ParallelWaitAll:
		sem := make(chan struct{}, runtime.NumCPU())
		ctx, cancel := context.WithCancel(context.Background())
		errChan := make(chan error, len(s.stores))
		for _, store := range s.stores {
			go func(store gokv.Store) {
				sem <- struct{}{}
				defer func() {
					<-sem
				}()
				select {
				case <-ctx.Done():
					// Don't call Set if ctx cancel was called,
					// which is the case when an error occurs in of the goroutines.
				case errChan <- store.Set(k, v):
				}
			}(store)
		}
		for i := 0; i < len(s.stores); i++ {
			if err := <-errChan; err != nil {
				// Stop remaining Set operations in goroutines immediately
				cancel()
				// Don't close errChan! A goroutine could still be in a Set call,
				// which would lead to sending an error to a closed channel, which would lead to a panic.
				return newMultiError(err)
			}
		}
		// All values are read from errChan, so we can safely close it
		close(errChan)
	case SequentialWaitFirst:
		err := s.stores[0].Set(k, v)
		if err != nil {
			return newMultiError(err)
		}
		go func(stores []gokv.Store) {
			for _, store := range stores {
				_ = store.Set(k, v) // Ignore errors
			}
		}(s.stores[1:])
	case SequentialWaitSuccess:
		fallthrough
	case SequentialWaitResult:
		i := 0
		multiError := MultiError{}
		for ; i < len(s.stores); i++ {
			store := s.stores[i]
			err := store.Set(k, v)
			if err != nil {
				multiError.addError(err)
			} else {
				// Success: Stop blocking iteration
				break // Note: i won't be incremented
			}
		}
		// Return now if all operations failed.
		if len(multiError.Errors) == len(s.stores) {
			return multiError
		}
		// Otherwise the recent operation was a success.
		// If more stores are left, go through them in a goroutine.
		nextIndex := i + 1
		if nextIndex <= len(s.stores) {
			go func(stores []gokv.Store) {
				for _, store := range stores {
					_ = store.Set(k, v) // Ignore errors
				}
			}(s.stores[nextIndex:])
		}
	default:
		return newMultiError(errors.New("The handling of the configured Set strategy is not implemented yet"))
	}
	return nil
}

// Get retrieves the stored value for the given key.
// You need to pass a pointer to the value, so in case of a struct
// the automatic unmarshalling can populate the fields of the object
// that v points to with the values of the retrieved object's values.
// The key must not be "" and the pointer must not be nil.
// The call is forwarded to the configured stores according to the configured strategy.
// Returned errors are of type `combiner.MultiError`
func (s Combiner) Get(k string, v interface{}) (bool, error) {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return false, newMultiError(err)
	}

	foundResult := false

	switch s.getStrategy {
	case SequentialWaitAll:
		prevFound := false
		var prevVal interface{}
		for i, store := range s.stores {
			found, err := store.Get(k, v)
			if err != nil {
				return found, newMultiError(err)
			}
			// No error, so compare with previous result.
			// (They must all be the same).
			if i > 0 {
				if !found && prevFound {
					// TODO: Error message should contain more detailed info,
					// for example *which* store contained a value and which didn't
					return found, newMultiError(errors.New("Value found in one store, but not in another"))
				} else if found && !reflect.DeepEqual(v, prevVal) {
					// TODO: Error message should contain more detailed info
					return found, newMultiError(errors.New("Value found in at least two stores, but they're not deeply equal"))
				}
			}
			prevFound = found
			prevVal = v
		}
		foundResult = prevFound
	case ParallelWaitAll:
		return foundResult, newMultiError(errors.New("Strategy ParallelWaitAll not implemented for Get yet"))
	case SequentialWaitFirst:
		var err error // For not having to use `:=` in the next step, which would lead to foundResult not being overwritten
		foundResult, err = s.stores[0].Get(k, v)
		if err != nil {
			return foundResult, newMultiError(err)
		}
	case SequentialWaitSuccess:
		multiError := MultiError{}
		var err error
		for _, store := range s.stores {
			foundResult, err = store.Get(k, v)
			if err != nil {
				multiError.addError(err)
			} else {
				// Success: Stop blocking iteration
				break
			}
		}
		// Return now if all operations failed.
		if len(multiError.Errors) == len(s.stores) {
			return false, multiError
		}
		// Otherwise the recent operation was a success.
	case SequentialWaitResult:
		multiError := MultiError{}
		var err error
		for _, store := range s.stores {
			foundResult, err = store.Get(k, v)
			if err != nil {
				multiError.addError(err)
			} else if foundResult {
				// Success: Stop blocking iteration
				break
			}
		}
		// Return now if all operations failed.
		if len(multiError.Errors) == len(s.stores) {
			return false, multiError
		}
		// Otherwise the recent operation was a success.
	default:
		return foundResult, newMultiError(errors.New("The handling of the configured Get strategy is not implemented yet"))
	}
	return foundResult, nil
}

// Delete deletes the stored value for the given key.
// Deleting a non-existing key-value pair does NOT lead to an error.
// The key must not be "".
// The call is forwarded to the configured stores according to the configured strategy.
// Returned errors are of type `combiner.MultiError`
func (s Combiner) Delete(k string) error {
	if err := util.CheckKey(k); err != nil {
		return err
	}

	switch s.deleteStrategy {
	case SequentialWaitAll:
		for _, store := range s.stores {
			err := store.Delete(k)
			if err != nil {
				return newMultiError(err)
			}
		}
	case ParallelWaitAll:
		sem := make(chan struct{}, runtime.NumCPU())
		ctx, cancel := context.WithCancel(context.Background())
		errChan := make(chan error, len(s.stores))
		for _, store := range s.stores {
			go func(store gokv.Store) {
				sem <- struct{}{}
				defer func() {
					<-sem
				}()
				select {
				case <-ctx.Done():
					// Don't call Delete if ctx cancel was called,
					// which is the case when an error occurs in of the goroutines.
				case errChan <- store.Delete(k):
				}
			}(store)
		}
		for i := 0; i < len(s.stores); i++ {
			if err := <-errChan; err != nil {
				// Stop remaining Delete operations in goroutines immediately
				cancel()
				// Don't close errChan! A goroutine could still be in a Delete call,
				// which would lead to sending an error to a closed channel, which would lead to a panic.
				return newMultiError(err)
			}
		}
		// All values are read from errChan, so we can safely close it
		close(errChan)
	case SequentialWaitFirst:
		err := s.stores[0].Delete(k)
		if err != nil {
			return newMultiError(err)
		}
		go func(stores []gokv.Store) {
			for _, store := range stores {
				_ = store.Delete(k) // Ignore errors
			}
		}(s.stores[1:])
	case SequentialWaitSuccess:
		fallthrough
	case SequentialWaitResult:
		i := 0
		multiError := MultiError{}
		for ; i < len(s.stores); i++ {
			store := s.stores[i]
			err := store.Delete(k)
			if err != nil {
				multiError.addError(err)
			} else {
				// Success: Stop blocking iteration
				break // Note: i won't be incremented
			}
		}
		// Return now if all operations failed.
		if len(multiError.Errors) == len(s.stores) {
			return multiError
		}
		// Otherwise the recent operation was a success.
		// If more stores are left, go through them in a goroutine.
		nextIndex := i + 1
		if nextIndex <= len(s.stores) {
			go func(stores []gokv.Store) {
				for _, store := range stores {
					_ = store.Delete(k) // Ignore errors
				}
			}(s.stores[nextIndex:])
		}
	default:
		return newMultiError(errors.New("The handling of the configured Delete strategy is not implemented yet"))
	}
	return nil
}

// Close closes all configured stores.
// The call is forwarded to the configured stores according to the configured strategy.
// Returned errors are of type `combiner.MultiError`
func (s Combiner) Close() error {
	switch s.closeStrategy {
	case SequentialWaitAll:
		for _, store := range s.stores {
			err := store.Close()
			if err != nil {
				return newMultiError(err)
			}
		}
	case ParallelWaitAll:
		sem := make(chan struct{}, runtime.NumCPU())
		ctx, cancel := context.WithCancel(context.Background())
		errChan := make(chan error, len(s.stores))
		for _, store := range s.stores {
			go func(store gokv.Store) {
				sem <- struct{}{}
				defer func() {
					<-sem
				}()
				select {
				case <-ctx.Done():
					// Don't call Close if ctx cancel was called,
					// which is the case when an error occurs in of the goroutines.
				case errChan <- store.Close():
				}
			}(store)
		}
		for i := 0; i < len(s.stores); i++ {
			if err := <-errChan; err != nil {
				// Stop remaining Close operations in goroutines immediately
				cancel()
				// Don't close errChan! A goroutine could still be in a Close call,
				// which would lead to sending an error to a closed channel, which would lead to a panic.
				return newMultiError(err)
			}
		}
		// All values are read from errChan, so we can safely close it
		close(errChan)
	case SequentialWaitFirst:
		err := s.stores[0].Close()
		if err != nil {
			return newMultiError(err)
		}
		go func(stores []gokv.Store) {
			for _, store := range stores {
				_ = store.Close() // Ignore errors
			}
		}(s.stores[1:])
	case SequentialWaitSuccess:
		fallthrough
	case SequentialWaitResult:
		i := 0
		multiError := MultiError{}
		for ; i < len(s.stores); i++ {
			store := s.stores[i]
			err := store.Close()
			if err != nil {
				multiError.addError(err)
			} else {
				// Success: Stop blocking iteration
				break // Note: i won't be incremented
			}
		}
		// Return now if all operations failed.
		if len(multiError.Errors) == len(s.stores) {
			return multiError
		}
		// Otherwise the recent operation was a success.
		// If more stores are left, go through them in a goroutine.
		nextIndex := i + 1
		if nextIndex <= len(s.stores) {
			go func(stores []gokv.Store) {
				for _, store := range stores {
					_ = store.Close() // Ignore errors
				}
			}(s.stores[nextIndex:])
		}
	default:
		return newMultiError(errors.New("The handling of the configured Close strategy is not implemented yet"))
	}
	return nil
}

// Options are the options for the combiner.
type Options struct {
	SetStrategy    Strategy
	GetStrategy    Strategy
	DeleteStrategy Strategy
	CloseStrategy  Strategy
}

// DefaultOptions is an Options object with default values.
// All operations' strategies are set to SequentialWaitAll, the most reliable one.
var DefaultOptions = Options{
	// SequentialWaitAll is iota, so 0, so the default value.
}

// NewCombiner creates a new Combiner.
// At least two stores must be passed.
// You should call `Close()` when you are done using it.
func NewCombiner(options Options, stores ...gokv.Store) (Combiner, error) {
	result := Combiner{}

	// Precondition check:
	// At least two stores
	if len(stores) < 2 {
		return result, errors.New("At least two stores must be passed for the creation of a combiner")
	}

	// No need to check for empty values in options, because the default value of a strategy is 0,
	// which is SequentialWaitAll, the one meant to be used as default.

	result.stores = stores
	result.setStrategy = options.SetStrategy
	result.getStrategy = options.GetStrategy
	result.deleteStrategy = options.DeleteStrategy
	result.closeStrategy = options.CloseStrategy
	return result, nil
}
