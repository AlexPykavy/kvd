package v2

import (
	"kvd/internal/store"
	"sync"
	"sync/atomic"
)

const (
	DefaultMyHashTableCapacity = 16
)

type MyEntry struct {
	key, value string

	overflow *MyEntry
}

type MyHashTable struct {
	shards    uint64
	muFactory func() store.RWLocker
	mu        []store.RWLocker

	n        atomic.Int64
	capacity uint64
	hasher   func(string) uint64
	entries  []*MyEntry
}

type MyHashTableOption func(*MyHashTable)

func WithCapacity(n uint64) MyHashTableOption {
	var c uint64 = 1
	for c < n {
		c <<= 1
	}

	return func(h *MyHashTable) {
		h.capacity = c
	}
}

func WithMutex(n uint64) MyHashTableOption {
	return func(h *MyHashTable) {
		h.shards = n
		h.muFactory = func() store.RWLocker { return &store.Mutex{} }
	}
}

func WithRWMutex(n uint64) MyHashTableOption {
	return func(h *MyHashTable) {
		h.shards = n
		h.muFactory = func() store.RWLocker { return &sync.RWMutex{} }
	}
}

func NewMyHashTable(options ...MyHashTableOption) *MyHashTable {
	h := &MyHashTable{
		shards:    1,
		muFactory: func() store.RWLocker { return &store.MutexStub{} },
		n:         atomic.Int64{},
		capacity:  DefaultMyHashTableCapacity,
		hasher: func(s string) uint64 {
			var h uint64 = 14695981039346656037

			for i := 0; i < len(s); i++ {
				h ^= uint64(s[i])
				h *= 1099511628211
			}

			return h
		},
	}

	for _, option := range options {
		option(h)
	}

	h.entries = make([]*MyEntry, h.capacity)
	h.mu = make([]store.RWLocker, h.shards)

	for i := range h.mu {
		h.mu[i] = h.muFactory()
	}

	return h
}

func (h *MyHashTable) IsThreadSafe() bool {
	if _, ok := h.mu[0].(*store.MutexStub); ok {
		return false
	}

	return true
}

func (h *MyHashTable) Put(key, value string) error {
	keyHash := h.hasher(key) % h.capacity
	shard := keyHash % h.shards

	h.mu[shard].Lock()

	pointer := &h.entries[keyHash]
	for *pointer != nil && (*pointer).key != key {
		pointer = &(*pointer).overflow
	}

	if *pointer != nil {
		(*pointer).value = value
	} else {
		*pointer = &MyEntry{
			key:   key,
			value: value,
		}
		h.n.Add(1)
	}

	h.mu[shard].Unlock()

	if h.n.Load() == int64(h.capacity) {
		h.rebalanceNaive()
	}

	return nil
}

func (h *MyHashTable) Get(key string) (string, error) {
	keyHash := h.hasher(key) % h.capacity
	shard := keyHash % h.shards

	h.mu[shard].RLock()
	defer h.mu[shard].RUnlock()

	pointer := h.entries[keyHash]
	for pointer != nil && pointer.key != key {
		pointer = pointer.overflow
	}

	var value string
	if pointer == nil {
		return value, store.ErrNotFound
	}

	return pointer.value, nil
}

func (h *MyHashTable) Delete(key string) error {
	keyHash := h.hasher(key) % h.capacity
	shard := keyHash % h.shards

	h.mu[shard].Lock()
	defer h.mu[shard].Unlock()

	pointer := &h.entries[keyHash]
	for *pointer != nil && (*pointer).key != key {
		pointer = &(*pointer).overflow
	}

	if *pointer == nil {
		return store.ErrNotFound
	}

	pointer = &(*pointer).overflow
	h.n.Add(-1)

	return nil
}

func (h *MyHashTable) Len() int {
	return int(h.n.Load())
}

// TODO: it's ineffecient to make a rebalancing under all locks
// - try using blue/green strategy, where all new elements are added
// to a new slice while old ones are gradually moved to the new one
// - Get() method must check both blue and green slices
// - swap blue and gree once the migration is complete
func (h *MyHashTable) rebalanceNaive() {
	for i := range h.mu {
		h.mu[i].Lock()
		defer h.mu[i].Unlock()
	}

	newCapacity := h.capacity << 1
	newEntries := make([]*MyEntry, newCapacity)

	for i := range h.entries {
		entry := h.entries[i]
		for entry != nil {
			newKeyHash := h.hasher(h.entries[i].key) % newCapacity

			pointer := &newEntries[newKeyHash]
			for *pointer != nil {
				pointer = &(*pointer).overflow
			}

			pointer = &entry

			entry = entry.overflow
		}
	}

	h.entries = newEntries
	h.capacity = newCapacity
}
