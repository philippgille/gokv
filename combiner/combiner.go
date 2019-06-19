package combiner

import (
	"runtime"
	"sync"

	"github.com/philippgille/gokv"
)

// Check we satisfy interface.
var _ gokv.Store = &Store{}

// Store is a `gokv.Store` implementation that forwards its
// calls to other `gokv.Store`s.
type Store struct {
	backends []gokv.Store
	parallel bool
}

// A helper type to check for errors when creating `gokv.Store`s.
type Backend struct {
	store gokv.Store
	err   error
}

// The options for this `gokv.Store` implementation.
type Options struct {
	Backends []Backend
	Parallel bool
}

// Returns a `Backend` that doesn't need to check for
// errors when it's being initialized.
func MustBackend(st gokv.Store) Backend {
	return Backend{store: st}
}

// Returns a `Backend` that may contain an error
// when it's initialized.
func NewBackend(st gokv.Store, err error) Backend {
	return Backend{st, err}
}

// Returns a new Store. At least two `gokv.Store`s are needed
// to initialize it. You should call `Close()` when you are done using it.
func NewStore(options Options) (*Store, error) {
	if len(options.Backends) < 2 {
		return nil, ErrNotEnoughStores
	}
	stores := make([]gokv.Store, len(options.Backends))
	for i := range options.Backends {
		if options.Backends[i].err != nil {
			return nil, options.Backends[i].err
		}
		stores[i] = options.Backends[i].store
	}
	s := &Store{
		backends: stores,
		parallel: options.Parallel,
	}
	return s, nil
}

// Implements `gokv.Store`. Forwards the call to each `gokv.Store`
// that was configured. Maximizes CPU usage if the parallel flag
// was set to true in the options.
func (s *Store) Set(k string, v interface{}) error {
	if s.parallel {
		return s.setParallel(k, v)
	}
	var errs []error
	for _, st := range s.backends {
		if err := st.Set(k, v); err != nil {
			errs = append(errs, err)
		}
	}
	if errs != nil {
		return newErrors(errs...)
	}
	return nil
}

func (s *Store) setParallel(k string, v interface{}) error {
	var wg sync.WaitGroup
	var errMux sync.Mutex
	var errs []error

	sem := make(chan struct{}, runtime.NumCPU())

	for _, st := range s.backends {
		wg.Add(1)
		go func(st gokv.Store) {
			sem <- struct{}{}
			defer func() {
				<-sem
				wg.Done()
			}()
			if err := st.Set(k, v); err != nil {
				errMux.Lock()
				errs = append(errs, err)
				errMux.Unlock()
			}
		}(st)
	}

	wg.Wait()

	if errs != nil {
		return newErrors(errs...)
	}
	return nil
}

// Implements `gokv.Store`. Returns the first successful call
// to `Get()`, between all the configured `gokv.Store`s.
func (s *Store) Get(k string, v interface{}) (bool, error) {
	var errs []error
	for _, st := range s.backends {
		if ok, err := st.Get(k, v); !ok {
			if err != nil {
				errs = append(errs, err)
			}
			continue
		}
		return true, nil
	}
	if errs != nil {
		return false, newErrors(errs...)
	}
	return false, nil
}

// Implements `gokv.Store`. Closes all the configured `gokv.Store`s.
func (s *Store) Close() error {
	var errs []error
	for _, st := range s.backends {
		if err := st.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if errs != nil {
		return newErrors(errs...)
	}
	return nil
}

// Implements `gokv.Store`. Forwards the call to each `gokv.Store`
// that was configured. Maximizes CPU usage if the parallel flag
// was set to true in the options.
func (s *Store) Delete(k string) error {
	if s.parallel {
		return s.delParallel(k)
	}
	var errs []error
	for _, st := range s.backends {
		if err := st.Delete(k); err != nil {
			errs = append(errs, err)
		}
	}
	if errs != nil {
		return newErrors(errs...)
	}
	return nil
}

func (s *Store) delParallel(k string) error {
	var wg sync.WaitGroup
	var errMux sync.Mutex
	var errs []error

	sem := make(chan struct{}, runtime.NumCPU())

	for _, st := range s.backends {
		wg.Add(1)
		go func(st gokv.Store) {
			sem <- struct{}{}
			defer func() {
				<-sem
				wg.Done()
			}()
			if err := st.Delete(k); err != nil {
				errMux.Lock()
				errs = append(errs, err)
				errMux.Unlock()
			}
		}(st)
	}

	wg.Wait()

	if errs != nil {
		return newErrors(errs...)
	}
	return nil
}
