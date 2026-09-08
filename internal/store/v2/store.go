package v2

import (
	"kvd/internal/store"
	"sync/atomic"
)

const (
	DefaultMyHashTableCapacity = 16
	DepthOffset                = 1000000000
)

type MyEntry struct {
	key, value string

	overflow *MyEntry
}

type MyHashTable struct {
	shards uint64
	mu     []store.RWLocker

	n        atomic.Int64
	capacity uint64
	hasher   func(string) uint64
	entries  []*MyEntry

	maxDepth   atomic.Uint64
	rebalances atomic.Uint64
}

func NewMyHashTable(options ...store.StoreConfigOption) *MyHashTable {
	cfg := &store.StoreConfig{
		Capacity:      DefaultMyHashTableCapacity,
		Shards:        1,
		LockerFactory: func() store.RWLocker { return &store.MutexStub{} },
	}

	for _, option := range options {
		option(cfg)
	}

	h := &MyHashTable{
		shards: cfg.Shards,
		mu:     make([]store.RWLocker, cfg.Shards),

		n:        atomic.Int64{},
		capacity: cfg.Capacity,
		hasher: func(s string) uint64 {
			var h uint64 = 14695981039346656037

			for i := 0; i < len(s); i++ {
				h ^= uint64(s[i])
				h *= 1099511628211
			}

			return h
		},
		entries: make([]*MyEntry, cfg.Capacity),
	}

	for i := range h.mu {
		h.mu[i] = cfg.LockerFactory()
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

	depth := h.capacity
	pointer := &h.entries[keyHash]
	for *pointer != nil && (*pointer).key != key {
		depth += DepthOffset
		pointer = &(*pointer).overflow
	}

	var created bool
	if *pointer != nil {
		(*pointer).value = value
	} else {
		*pointer = &MyEntry{
			key:   key,
			value: value,
		}
		created = true
	}

	h.mu[shard].Unlock()

	for {
		currentMaxDepth := h.maxDepth.Load()
		if depth <= currentMaxDepth || h.maxDepth.CompareAndSwap(currentMaxDepth, depth) {
			break
		}
	}

	// There are several issues with the current rebalancing implementation:
	// 1. There can be multiple rebalancing happening at the same time, resulting in race condition and lost entries
	// as every .rebalanceNaive() call creates a local slice to keep elements
	//
	// 2. There is still an issue with triggering the rabalancing as .rebalanceNaive() there is a chance
	// that there will be so many concurrent puts that it will jump over the doubled h.capacity
	// before the rebalancing actually aquires all the locks and as a result the rebalancing will be skipped
	//
	// 3. It's ineffecient to make a rebalancing under all locks as it blocks the readers.
	// We could try using blue/green strategy, where all new entries are added to a new slice
	// while old are gradually moved to the new one and the .Get() method checks both blue and green slices
	//

	// I see several options to mitigate the above issues:
	// - rebalancingMu sync.RWMutex and RLock() during operations and Lock() during rebalancing might help
	// - extend blue/green approach, so a MyHashTable has a list of MyStaticHashTable:
	//   a. once rebalancing is needed, we just append an increased MyStaticHashTable to the front
	//      and use it before the subsequent
	//   b. we can also compact the shards on some operation path, e.g. Get
	//   c. when all shards become migrated, the MyStaticHashTable is deleted from the list
	if created && h.n.Add(1) == int64(h.capacity) {
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

	*pointer = (*pointer).overflow
	h.n.Add(-1)

	return nil
}

func (h *MyHashTable) Len() int {
	return int(h.n.Load())
}

func (h *MyHashTable) MaxDepth() (uint32, uint32) {
	return uint32(h.maxDepth.Load() / DepthOffset), uint32(h.maxDepth.Load() % DepthOffset)
}

func (h *MyHashTable) Rebalances() uint64 {
	return h.rebalances.Load()
}

func (h *MyHashTable) rebalanceNaive() {
	h.rebalances.Add(1)

	for i := range h.mu {
		h.mu[i].Lock()
		defer h.mu[i].Unlock()
	}

	newCapacity := h.capacity << 1
	newEntries := make([]*MyEntry, newCapacity)

	for i := range h.entries {
		for h.entries[i] != nil {
			newKeyHash := h.hasher(h.entries[i].key) % newCapacity

			pointer := &newEntries[newKeyHash]
			for *pointer != nil {
				pointer = &(*pointer).overflow
			}

			*pointer = h.entries[i]
			h.entries[i] = h.entries[i].overflow
			(*pointer).overflow = nil
		}
	}

	h.entries = newEntries
	h.capacity = newCapacity
}
