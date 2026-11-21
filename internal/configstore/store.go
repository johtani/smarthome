// Package configstore provides atomic access to runtime-reloadable configuration snapshots.
package configstore

import (
	"sync"
	"sync/atomic"

	"github.com/johtani/smarthome/subcommand"
)

// Store provides lock-free access to immutable configuration snapshots.
type Store struct {
	value     atomic.Value
	mu        sync.Mutex
	nextID    uint64
	observers map[uint64]func(subcommand.Config)
}

// New creates a Store initialized with the given config snapshot.
func New(initial subcommand.Config) *Store {
	s := &Store{}
	s.value.Store(initial)
	return s
}

// Get returns the current config snapshot.
func (s *Store) Get() subcommand.Config {
	return s.value.Load().(subcommand.Config)
}

// Set atomically replaces the current config snapshot.
func (s *Store) Set(cfg subcommand.Config) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.value.Store(cfg)
	for _, observer := range s.observers {
		observer(cfg)
	}
}

// Observe delivers the current snapshot and each replacement in publication order.
// The callback must not modify the store. Call the returned function to unsubscribe.
func (s *Store) Observe(observer func(subcommand.Config)) func() {
	s.mu.Lock()
	defer s.mu.Unlock()
	observer(s.Get())
	if s.observers == nil {
		s.observers = make(map[uint64]func(subcommand.Config))
	}
	id := s.nextID
	s.nextID++
	s.observers[id] = observer
	return func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		delete(s.observers, id)
	}
}

// Reload publishes a new snapshot only when loading and validation succeed.
func (s *Store) Reload(load func() (subcommand.Config, error)) error {
	cfg, err := load()
	if err != nil {
		return err
	}
	s.Set(cfg)
	return nil
}
