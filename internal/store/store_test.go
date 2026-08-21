package store_test

import (
	"fmt"
	"kvd/internal/store"
	v0 "kvd/internal/store/v0"
	"sync/atomic"
	"testing"
)

func createAllStores() []struct {
	name  string
	store store.Store
} {
	return []struct {
		name  string
		store store.Store
	}{
		{"v0", v0.NewStore()},
	}
}

func TestNewStore(t *testing.T) {
	t.Parallel()

	tests := createAllStores()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if test.store.Len() != 0 {
				t.Errorf("Just initialized s.Len() = %d, want 0", test.store.Len())
			}
		})
	}
}

func TestStoreGetNotExisting(t *testing.T) {
	t.Parallel()

	key := "my-key"
	tests := createAllStores()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if value, err := test.store.Get(key); value != "" || err != store.ErrNotFound {
				t.Errorf("s.Get(%q) = %q, %v, want %q, %v", key, value, "", err, store.ErrNotFound)
			}
		})
	}
}

func TestStorePut(t *testing.T) {
	t.Parallel()

	key, value := "my-key", "my-value"

	tests := createAllStores()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if err := test.store.Put(key, value); err != nil {
				t.Errorf("s.Put(%q, %q) = %v, want %v", key, value, err, nil)
			}

			if test.store.Len() != 1 {
				t.Errorf("s.Len() = %d, want 1", test.store.Len())
			}

			if obtained, err := test.store.Get(key); obtained != value || err != nil {
				t.Errorf("s.Get(%q) = %q, %v, want %q, %v", key, obtained, err, value, nil)
			}
		})
	}
}

func TestStoreDeleteNotExisting(t *testing.T) {
	t.Parallel()

	key := "my-key"
	tests := createAllStores()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if err := test.store.Delete(key); err != store.ErrNotFound {
				t.Errorf("s.Delete(%q) = %v, want %v", key, err, store.ErrNotFound)
			}
		})
	}
}

func TestStoreDelete(t *testing.T) {
	t.Parallel()

	key, value := "my-key", "my-value"

	tests := createAllStores()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if err := test.store.Put(key, value); err != nil {
				t.Errorf("s.Put(%q, %q) = %v, want %v", key, value, err, nil)
			}

			if test.store.Len() != 1 {
				t.Errorf("s.Len() = %d, want 1", test.store.Len())
			}

			if obtained, err := test.store.Get(key); obtained != value || err != nil {
				t.Errorf("s.Get(%q) = %q, %v, want %q, %v", key, obtained, err, value, nil)
			}

			if err := test.store.Delete(key); err != nil {
				t.Errorf("s.Delete(%q) = %v, want %v", key, err, nil)
			}

			if test.store.Len() != 0 {
				t.Errorf("s.Len() = %d, want 0", test.store.Len())
			}
		})
	}
}

func BenchmarkStorePutSequentially(b *testing.B) {
	tests := createAllStores()

	for _, test := range tests {
		b.Run(test.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				key := fmt.Sprint(i)
				value := key

				test.store.Put(key, value)
			}

			if test.store.Len() != b.N {
				b.Errorf("s.Len() = %d, want %d", test.store.Len(), b.N)
			}
		})
	}
}

func BenchmarkStorePutSequentiallyWithAtomicCounter(b *testing.B) {
	tests := createAllStores()

	for _, test := range tests {
		b.Run(test.name, func(b *testing.B) {
			var counter atomic.Uint64
			for i := 0; i < b.N; i++ {
				key := fmt.Sprint(counter.Add(1))
				value := key

				test.store.Put(key, value)
			}

			if test.store.Len() != b.N {
				b.Errorf("s.Len() = %d, want %d", test.store.Len(), b.N)
			}
		})
	}
}

func BenchmarkStorePutConcurrently(b *testing.B) {
	tests := createAllStores()

	for _, test := range tests {
		b.Run(test.name, func(b *testing.B) {
			var counter atomic.Uint64
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					key := fmt.Sprint(counter.Add(1))
					value := key

					test.store.Put(key, value)
				}
			})

			if test.store.Len() != b.N {
				b.Errorf("s.Len() = %d, want %d", test.store.Len(), b.N)
			}
		})
	}
}

func BenchmarkStorePutTheSame(b *testing.B) {
	key, value := "my-key", "my-value"
	tests := createAllStores()

	for _, test := range tests {
		b.Run(test.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				test.store.Put(key, value)
			}

			if test.store.Len() != 1 {
				b.Errorf("s.Len() = %d, want %d", test.store.Len(), 1)
			}
		})
	}
}

func BenchmarkStorePutTheSameConcurrently(b *testing.B) {
	key, value := "my-key", "my-value"
	tests := createAllStores()

	for _, test := range tests {
		b.Run(test.name, func(b *testing.B) {
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					test.store.Put(key, value)
				}
			})

			if test.store.Len() != 1 {
				b.Errorf("s.Len() = %d, want %d", test.store.Len(), 1)
			}
		})
	}
}
