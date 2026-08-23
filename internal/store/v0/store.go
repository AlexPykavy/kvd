package v0

import "kvd/internal/store"

type Store struct {
	m map[string]string
}

type StoreOption func(*Store)

func WithCapacity(n int) StoreOption {
	return func(s *Store) {
		s.m = make(map[string]string, n)
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

	return s
}

func (s *Store) IsThreadSafe() bool {
	return false
}

func (s *Store) Put(key, value string) error {
	s.m[key] = value

	return nil
}

func (s *Store) Get(key string) (string, error) {
	var value string
	if value, ok := s.m[key]; ok {
		return value, nil
	}

	return value, store.ErrNotFound
}

func (s *Store) Delete(key string) error {
	if _, ok := s.m[key]; ok {
		delete(s.m, key)

		return nil
	}

	return store.ErrNotFound
}

func (s *Store) Len() int {
	return len(s.m)
}
