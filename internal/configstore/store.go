// Package configstore provides atomic access to runtime-reloadable configuration snapshots.
package configstore

import (
	"sync/atomic"

	"github.com/johtani/smarthome/subcommand"
)

// Store provides lock-free access to immutable configuration snapshots.
type Store struct {
	value atomic.Value
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
	s.value.Store(cfg)
}
