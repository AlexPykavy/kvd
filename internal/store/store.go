package store

import "errors"

var (
	ErrNotFound = errors.New("object not found")
)

type Store interface {
	Put(key, value string) error
	Get(key string) (string, error)
	Delete(key string) error
	Len() int
}
