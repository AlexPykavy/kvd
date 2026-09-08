package v3

import (
	"container/list"
	"kvd/internal/store"
	"sync"
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

	hasher func(s string) uint64
	hts    list.List
	htsMu  sync.RWMutex

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

		hasher: func(s string) uint64 {
			var h uint64 = 14695981039346656037

			for i := 0; i < len(s); i++ {
				h ^= uint64(s[i])
				h *= 1099511628211
			}

			return h
		},
	}

	for i := range h.mu {
		h.mu[i] = cfg.LockerFactory()
	}

	h.hts.PushFront(NewMyStaticHashTable(cfg.Capacity, h.hasher))

	return h
}

func (h *MyHashTable) IsThreadSafe() bool {
	if _, ok := h.mu[0].(*store.MutexStub); ok {
		return false
	}

	return true
}

func (h *MyHashTable) Put(key, value string) error {
	shard := h.hasher(key) % h.shards

	h.htsMu.RLock()
	frontHt := h.hts.Front().Value.(*MyStaticHashTable)
	h.htsMu.RUnlock()

	h.mu[shard].Lock()

	frontEntry, depth := frontHt.Find(key)
	if *frontEntry != nil {
		(*frontEntry).value = value

		h.mu[shard].Unlock()

		return nil
	}

	// maybe it's better to move all shard entries to the frontHt, not the only one??
	var entry **MyEntry
	for e := h.hts.Front().Next(); e != nil; e = e.Next() {
		ht := e.Value.(*MyStaticHashTable)

		if entry, _ = ht.Find(key); *entry != nil {
			*entry = (*entry).overflow

			if atomic.AddInt64(&ht.n, -1) == 0 {
				h.htsMu.Lock()
				h.hts.Remove(e)
				h.htsMu.Unlock()
			}

			break
		}
	}

	*frontEntry = &MyEntry{
		key:   key,
		value: value,
	}

	h.mu[shard].Unlock()

	depth += DepthOffset

	if atomic.AddInt64(&frontHt.n, 1) == int64(frontHt.capacity) {
		h.htsMu.Lock()
		// this seems to be required to protect from decreasing the capacity
		// if another goroutine already succeeded to increase it
		if newFrontHt := h.hts.Front().Value.(*MyStaticHashTable); newFrontHt.capacity == frontHt.capacity {
			h.hts.PushFront(NewMyStaticHashTable(frontHt.capacity<<1, h.hasher))
		}
		h.htsMu.Unlock()
	}

	for {
		currentMaxDepth := h.maxDepth.Load()
		if depth <= currentMaxDepth || h.maxDepth.CompareAndSwap(currentMaxDepth, depth) {
			break
		}
	}

	return nil
}

func (h *MyHashTable) Get(key string) (string, error) {
	shard := h.hasher(key) % h.shards

	h.mu[shard].RLock()
	defer h.mu[shard].RUnlock()

	var value string
	for e := h.hts.Front(); e != nil; e = e.Next() {
		if entry, _ := e.Value.(*MyStaticHashTable).Find(key); *entry != nil {
			return (*entry).value, nil
		}
	}

	return value, store.ErrNotFound
}

func (h *MyHashTable) Delete(key string) error {
	shard := h.hasher(key) % h.shards

	h.mu[shard].Lock()

	for e := h.hts.Front(); e != nil; e = e.Next() {
		ht := e.Value.(*MyStaticHashTable)

		if entry, _ := ht.Find(key); *entry != nil {
			*entry = (*entry).overflow

			h.mu[shard].Unlock()

			if atomic.AddInt64(&ht.n, -1) == 0 {
				h.htsMu.Lock()
				h.hts.Remove(e)
				h.htsMu.Unlock()
			}

			return nil
		}
	}

	h.mu[shard].Unlock()

	return store.ErrNotFound
}

func (h *MyHashTable) Len() int {
	n := 0

	h.htsMu.RLock()

	for e := h.hts.Front(); e != nil; e = e.Next() {
		ht := e.Value.(*MyStaticHashTable)

		n += int(atomic.LoadInt64(&ht.n))
	}

	h.htsMu.RUnlock()

	return n
}

func (h *MyHashTable) MaxDepth() (uint32, uint32) {
	return uint32(h.maxDepth.Load() / DepthOffset), uint32(h.maxDepth.Load() % DepthOffset)
}

func (h *MyHashTable) Rebalances() uint64 {
	return h.rebalances.Load()
}
