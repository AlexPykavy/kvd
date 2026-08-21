package v0

import "kvd/internal/store"

type Store struct {
	m map[string]string
}

func NewStore() *Store {
	return &Store{
		m: make(map[string]string),
	}
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
