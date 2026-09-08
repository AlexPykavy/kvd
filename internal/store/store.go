package store

import (
	"errors"
	"sync"
)

var (
	ErrNotFound = errors.New("object not found")
)

type StoreConfig struct {
	Capacity      uint64
	Shards        uint64
	LockerFactory func() RWLocker
}

type StoreConfigOption func(*StoreConfig)

func WithCapacity(n uint64) StoreConfigOption {
	var c uint64 = 1
	for c < n {
		c <<= 1
	}

	return func(cfg *StoreConfig) {
		cfg.Capacity = c
	}
}

func WithMutex(n uint64) StoreConfigOption {
	var s uint64 = 1
	for s < n {
		s <<= 1
	}

	return func(cfg *StoreConfig) {
		cfg.Shards = s
		cfg.LockerFactory = func() RWLocker { return &Mutex{} }
	}
}

func WithRWMutex(n uint64) StoreConfigOption {
	var s uint64 = 1
	for s < n {
		s <<= 1
	}

	return func(cfg *StoreConfig) {
		cfg.Shards = s
		cfg.LockerFactory = func() RWLocker { return &sync.RWMutex{} }
	}
}

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
