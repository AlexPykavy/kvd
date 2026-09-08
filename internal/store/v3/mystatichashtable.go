package v3

type MyStaticHashTable struct {
	n        int64
	capacity uint64
	hasher   func(string) uint64
	entries  []*MyEntry
}

type MyStaticHashTableOption func(*MyStaticHashTable)

func NewMyStaticHashTable(n uint64, hasher func(string) uint64) *MyStaticHashTable {
	return &MyStaticHashTable{
		n:        0,
		capacity: n,
		hasher:   hasher,
		entries:  make([]*MyEntry, n),
	}
}

func (h *MyStaticHashTable) Find(key string) (**MyEntry, uint64) {
	keyHash := h.hasher(key) % h.capacity

	depth := h.capacity
	pointer := &h.entries[keyHash]
	for *pointer != nil && (*pointer).key != key {
		depth += DepthOffset
		pointer = &(*pointer).overflow
	}

	return pointer, depth
}
