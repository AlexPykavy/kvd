package v1

import (
	"kvd/internal/store"
	"sync"
)

type Store struct {
	mu store.RWLocker

	m map[string]string
}

type StoreOption func(*Store)

func WithCapacity(n int) StoreOption {
	return func(s *Store) {
		s.m = make(map[string]string, n)
	}
}

func WithMutex() StoreOption {
	return func(s *Store) {
		s.mu = &store.Mutex{}
	}
}

func WithRWMutex() StoreOption {
	return func(s *Store) {
		s.mu = &sync.RWMutex{}
	}
}

func NewStore(options ...StoreOption) *Store {
	s := &Store{}

	for _, option := range options {
		option(s)
	}

	if s.m == nil {
		s.m = make(map[string]string)
	}

	if s.mu == nil {
		s.mu = &store.MutexStub{}
	}

	return s
}

func (s *Store) IsThreadSafe() bool {
	if _, ok := s.mu.(*store.MutexStub); ok {
		return false
	}

	return true
}

func (s *Store) Put(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.m[key] = value

	return nil
}

func (s *Store) Get(key string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var value string
	if value, ok := s.m[key]; ok {
		return value, nil
	}

	return value, store.ErrNotFound
}

func (s *Store) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.m[key]; ok {
		delete(s.m, key)

		return nil
	}

	return store.ErrNotFound
}

func (s *Store) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.m)
}
