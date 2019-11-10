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

// Combiner is a `gokv.Store` implementation that forwards its
// calls to other `gokv.Store`s with configurable strategies.
// At least two stores must be set.
type Combiner struct {
	stores         []gokv.Store
	setStrategy    UpdateStrategy
	getStrategy    GetStrategy
	deleteStrategy UpdateStrategy
	closeStrategy  CloseStrategy
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
	case UpdateSequentialWaitAll:
		multiError := MultiError{}
		for _, store := range s.stores {
			err := store.Set(k, v)
			if err != nil {
				multiError.addError(err)
			}
		}
		if len(multiError.Errors) > 0 {
			return multiError
		}
	case UpdateParallelWaitAll:
		multiError := MultiError{}
		sem := make(chan struct{}, runtime.NumCPU())
		errChan := make(chan error, len(s.stores))
		for _, store := range s.stores {
			go func(store gokv.Store) {
				sem <- struct{}{}
				defer func() {
					<-sem
				}()
				errChan <- store.Set(k, v)
				return
			}(store)
		}
		for i := 0; i < len(s.stores); i++ {
			if err := <-errChan; err != nil {
				multiError.addError(err)
			}
		}
		// All values are read from errChan, so we can safely close it
		close(errChan)
		if len(multiError.Errors) > 0 {
			return multiError
		}
	case UpdateSequentialWaitErrorThenContinue:
		multiError := MultiError{}
		continueFrom := 0
		for i, store := range s.stores {
			err := store.Set(k, v)
			if err != nil {
				multiError.addError(err)
				// Stop the blocking calls, continue later in a goroutine
				continueFrom = i + 1
				break
			}
		}
		// Continue any remaining operations in a goroutine.
		// `continueFrom == 0` would mean there was no error.
		// `continueFrom == len(s.stores)` would mean the error occurred for the last store.
		if continueFrom != 0 && continueFrom < len(s.stores) {
			go func(stores []gokv.Store) {
				for _, store := range stores {
					_ = store.Set(k, v) // Ignore errors
				}
				return
			}(s.stores[continueFrom:])
		}
		if len(multiError.Errors) > 0 {
			return multiError
		}
	case UpdateParallelWaitErrorThenContinue:
		sem := make(chan struct{}, runtime.NumCPU())
		errChan := make(chan error, len(s.stores))
		for _, store := range s.stores {
			go func(store gokv.Store) {
				sem <- struct{}{}
				defer func() {
					<-sem
				}()
				errChan <- store.Set(k, v)
				return
			}(store)
		}
		for i := 0; i < len(s.stores); i++ {
			if err := <-errChan; err != nil {
				// Just return upon the first error,
				// all remaining ops are executed in goroutines already.
				//
				// Don't close errChan! A goroutine could still be in a Set call,
				// which would lead to sending an error to a closed channel, which would lead to a panic.
				return newMultiError(err)
			}
		}
		// All values are read from errChan, so we can safely close it
		close(errChan)
	case UpdateSequentialWaitErrorThenSkip:
		for _, store := range s.stores {
			err := store.Set(k, v)
			if err != nil {
				return newMultiError(err)
			}
		}
	case UpdateParallelWaitErrorThenSkip:
		sem := make(chan struct{}, runtime.NumCPU())
		ctx, cancel := context.WithCancel(context.Background())
		errChan := make(chan error, len(s.stores))
		for _, store := range s.stores {
			go func(store gokv.Store) {
				sem <- struct{}{}
				defer func() {
					<-sem
				}()
				// Don't call Set if ctx cancel was called,
				// which is the case when an error occurs in of the goroutines.
				select {
				case <-ctx.Done():
				default:
					errChan <- store.Set(k, v)
				}
				return
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
	case UpdateSequentialWaitNoError:
		multiError := MultiError{}
		continueFrom := 0
		for i, store := range s.stores {
			err := store.Set(k, v)
			if err != nil {
				multiError.addError(err)
			} else {
				// Success. Stop blocking iteration!
				continueFrom = i + 1
				break
			}
		}
		// Return now if all operations failed.
		if len(multiError.Errors) == len(s.stores) {
			return multiError
		}
		// Otherwise the recent operation was a success.
		// If more stores are left, go through them in a goroutine.
		if continueFrom <= len(s.stores) {
			go func(stores []gokv.Store) {
				for _, store := range stores {
					_ = store.Set(k, v) // Ignore errors
				}
				return
			}(s.stores[continueFrom:])
		}
	case UpdateSequentialWaitFirst:
		err := s.stores[0].Set(k, v)
		if err != nil {
			return newMultiError(err)
		}
		go func(stores []gokv.Store) {
			for _, store := range stores {
				_ = store.Set(k, v) // Ignore errors
			}
			return
		}(s.stores[1:])
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
	case GetSequentialWaitError:
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
			prevVal = v // TODO: Probably requires deep copying
		}
		foundResult = prevFound
	case GetSequentialWaitValue:
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
	case GetSequentialWaitNoError:
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
	case GetSequentialWaitFirst:
		var err error // For not having to use `:=` in the next step, which would lead to foundResult not being overwritten
		foundResult, err = s.stores[0].Get(k, v)
		if err != nil {
			return foundResult, newMultiError(err)
		}
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
	case UpdateSequentialWaitAll:
		multiError := MultiError{}
		for _, store := range s.stores {
			err := store.Delete(k)
			if err != nil {
				multiError.addError(err)
			}
		}
		if len(multiError.Errors) > 0 {
			return multiError
		}
	case UpdateParallelWaitAll:
		multiError := MultiError{}
		sem := make(chan struct{}, runtime.NumCPU())
		errChan := make(chan error, len(s.stores))
		for _, store := range s.stores {
			go func(store gokv.Store) {
				sem <- struct{}{}
				defer func() {
					<-sem
				}()
				errChan <- store.Delete(k)
				return
			}(store)
		}
		for i := 0; i < len(s.stores); i++ {
			if err := <-errChan; err != nil {
				multiError.addError(err)
			}
		}
		// All values are read from errChan, so we can safely close it
		close(errChan)
		if len(multiError.Errors) > 0 {
			return multiError
		}
	case UpdateSequentialWaitErrorThenContinue:
		multiError := MultiError{}
		continueFrom := 0
		for i, store := range s.stores {
			err := store.Delete(k)
			if err != nil {
				multiError.addError(err)
				// Stop the blocking calls, continue later in a goroutine
				continueFrom = i + 1
				break
			}
		}
		// Continue any remaining operations in a goroutine.
		// `continueFrom == 0` would mean there was no error.
		// `continueFrom == len(s.stores)` would mean the error occurred for the last store.
		if continueFrom != 0 && continueFrom < len(s.stores) {
			go func(stores []gokv.Store) {
				for _, store := range stores {
					_ = store.Delete(k) // Ignore errors
				}
				return
			}(s.stores[continueFrom:])
		}
		if len(multiError.Errors) > 0 {
			return multiError
		}
	case UpdateParallelWaitErrorThenContinue:
		sem := make(chan struct{}, runtime.NumCPU())
		errChan := make(chan error, len(s.stores))
		for _, store := range s.stores {
			go func(store gokv.Store) {
				sem <- struct{}{}
				defer func() {
					<-sem
				}()
				errChan <- store.Delete(k)
				return
			}(store)
		}
		for i := 0; i < len(s.stores); i++ {
			if err := <-errChan; err != nil {
				// Just return upon the first error,
				// all remaining ops are executed in goroutines already.
				//
				// Don't close errChan! A goroutine could still be in a Set call,
				// which would lead to sending an error to a closed channel, which would lead to a panic.
				return newMultiError(err)
			}
		}
		// All values are read from errChan, so we can safely close it
		close(errChan)
	case UpdateSequentialWaitErrorThenSkip:
		for _, store := range s.stores {
			err := store.Delete(k)
			if err != nil {
				return newMultiError(err)
			}
		}
	case UpdateParallelWaitErrorThenSkip:
		sem := make(chan struct{}, runtime.NumCPU())
		ctx, cancel := context.WithCancel(context.Background())
		errChan := make(chan error, len(s.stores))
		for _, store := range s.stores {
			go func(store gokv.Store) {
				sem <- struct{}{}
				defer func() {
					<-sem
				}()
				// Don't call Set if ctx cancel was called,
				// which is the case when an error occurs in of the goroutines.
				select {
				case <-ctx.Done():
				default:
					errChan <- store.Delete(k)
				}
				return
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
	case UpdateSequentialWaitNoError:
		multiError := MultiError{}
		continueFrom := 0
		for i, store := range s.stores {
			err := store.Delete(k)
			if err != nil {
				multiError.addError(err)
			} else {
				// Success. Stop blocking iteration!
				continueFrom = i + 1
				break
			}
		}
		// Return now if all operations failed.
		if len(multiError.Errors) == len(s.stores) {
			return multiError
		}
		// Otherwise the recent operation was a success.
		// If more stores are left, go through them in a goroutine.
		if continueFrom <= len(s.stores) {
			go func(stores []gokv.Store) {
				for _, store := range stores {
					_ = store.Delete(k) // Ignore errors
				}
				return
			}(s.stores[continueFrom:])
		}
	case UpdateSequentialWaitFirst:
		err := s.stores[0].Delete(k)
		if err != nil {
			return newMultiError(err)
		}
		go func(stores []gokv.Store) {
			for _, store := range stores {
				_ = store.Delete(k) // Ignore errors
			}
			return
		}(s.stores[1:])
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
	case CloseSequentialWaitAll:
		multiError := MultiError{}
		for _, store := range s.stores {
			err := store.Close()
			if err != nil {
				multiError.addError(err)
			}
		}
		if len(multiError.Errors) > 0 {
			return multiError
		}
	case CloseParallelWaitAll:
		multiError := MultiError{}
		sem := make(chan struct{}, runtime.NumCPU())
		errChan := make(chan error, len(s.stores))
		for _, store := range s.stores {
			go func(store gokv.Store) {
				sem <- struct{}{}
				defer func() {
					<-sem
				}()
				errChan <- store.Close()
				return
			}(store)
		}
		for i := 0; i < len(s.stores); i++ {
			if err := <-errChan; err != nil {
				multiError.addError(err)
			}
		}
		// All values are read from errChan, so we can safely close it
		close(errChan)
		if len(multiError.Errors) > 0 {
			return multiError
		}
	default:
		return newMultiError(errors.New("The handling of the configured Close strategy is not implemented yet"))
	}
	return nil
}

// Options are the options for the combiner.
type Options struct {
	SetStrategy    UpdateStrategy
	GetStrategy    GetStrategy
	DeleteStrategy UpdateStrategy
	CloseStrategy  CloseStrategy
}

// DefaultOptions is an Options object with default values.
var DefaultOptions = Options{
	SetStrategy:    UpdateParallelWaitAll,
	GetStrategy:    GetSequentialWaitError,
	DeleteStrategy: UpdateParallelWaitAll,
	CloseStrategy:  CloseParallelWaitAll,
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

	// Set default values if necessary.
	// Even if no stategy with int value 0 exists, 0 can still be used.
	if options.SetStrategy == 0 {
		options.SetStrategy = DefaultOptions.SetStrategy
	}
	if options.GetStrategy == 0 {
		options.GetStrategy = DefaultOptions.GetStrategy
	}
	if options.DeleteStrategy == 0 {
		options.DeleteStrategy = DefaultOptions.DeleteStrategy
	}
	if options.CloseStrategy == 0 {
		options.CloseStrategy = DefaultOptions.CloseStrategy
	}

	result.stores = stores
	result.setStrategy = options.SetStrategy
	result.getStrategy = options.GetStrategy
	result.deleteStrategy = options.DeleteStrategy
	result.closeStrategy = options.CloseStrategy
	return result, nil
}
