package store

import "errors"

var (
	ErrNotFound = errors.New("object not found")
)

type Store interface {
	IsThreadSafe() bool

	Put(key, value string) error
	Get(key string) (string, error)
	Delete(key string) error
	Len() int
}

type MyHashTableDebug interface {
	MaxDepth() (maxDepth uint32, capacity uint32)
	Rebalances() uint64
}
