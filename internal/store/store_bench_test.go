package store_test

import (
	"fmt"
	"kvd/internal/store"
	v0 "kvd/internal/store/v0"
	"os"
	"sync/atomic"
	"testing"
)

var keys [10_000_000]string

func TestMain(m *testing.M) {
	for i := range keys {
		keys[i] = fmt.Sprint(i)
	}

	os.Exit(m.Run())
}

func createAllBenchmarkStores() []struct {
	name  string
	store store.Store
} {
	return []struct {
		name  string
		store store.Store
	}{
		{"v0", v0.NewStore()},
		{"v0withCapacity1000", v0.NewStore(v0.WithCapacity(1000))},
		{"v0withCapacity1000_000", v0.NewStore(v0.WithCapacity(1000_000))},
		{"v0withCapacity10_000_000", v0.NewStore(v0.WithCapacity(10_000_000))},
	}
}

func BenchmarkStorePutSequentially(b *testing.B) {
	tests := createAllBenchmarkStores()

	for _, test := range tests {
		b.Run(test.name, func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				test.store.Put(keys[i], keys[i])
			}

			if test.store.Len() != b.N {
				b.Errorf("s.Len() = %d, want %d", test.store.Len(), b.N)
			}
		})
	}
}

func BenchmarkStoreGetSequentially(b *testing.B) {
	tests := createAllBenchmarkStores()

	for _, test := range tests {
		b.Run(test.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				test.store.Put(keys[i], keys[i])
			}

			if test.store.Len() != b.N {
				b.Errorf("s.Len() = %d, want %d", test.store.Len(), b.N)
			}

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				test.store.Get(keys[i])
			}
		})
	}
}

func BenchmarkStorePutSequentiallyWithAtomicCounter(b *testing.B) {
	tests := createAllBenchmarkStores()

	for _, test := range tests {
		b.Run(test.name, func(b *testing.B) {
			b.ResetTimer()

			var counter atomic.Int64
			for i, n := counter.Load(), int64(b.N); i < n; i = counter.Add(1) {
				test.store.Put(keys[i], keys[i])
			}

			if test.store.Len() != b.N {
				b.Errorf("s.Len() = %d, want %d", test.store.Len(), b.N)
			}
		})
	}
}

func BenchmarkStorePutConcurrently(b *testing.B) {
	tests := createAllBenchmarkStores()

	for _, test := range tests {
		if !test.store.IsThreadSafe() {
			b.Skipf("skipping test because the store is not thread-safe and we will get %q", "fatal error: concurrent map writes")
		}

		b.Run(test.name, func(b *testing.B) {
			b.ResetTimer()

			var counter atomic.Uint64
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					key := keys[counter.Add(1)]

					test.store.Put(key, key)
				}
			})

			if test.store.Len() != b.N {
				b.Errorf("s.Len() = %d, want %d", test.store.Len(), b.N)
			}
		})
	}
}

func BenchmarkStoreGetConcurrently(b *testing.B) {
	tests := createAllBenchmarkStores()

	for _, test := range tests {
		if !test.store.IsThreadSafe() {
			b.Skipf("skipping test because the store is not thread-safe and we will get %q", "fatal error: concurrent map writes")
		}

		b.Run(test.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				test.store.Put(keys[i], keys[i])
			}

			if test.store.Len() != b.N {
				b.Errorf("s.Len() = %d, want %d", test.store.Len(), b.N)
			}

			b.ResetTimer()

			var counter atomic.Uint64
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					test.store.Get(keys[counter.Add(1)])
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
	tests := createAllBenchmarkStores()

	for _, test := range tests {
		b.Run(test.name, func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				test.store.Put(key, value)
			}

			if test.store.Len() != 1 {
				b.Errorf("s.Len() = %d, want %d", test.store.Len(), 1)
			}
		})
	}
}

func BenchmarkStoreGetTheSame(b *testing.B) {
	key, value := "my-key", "my-value"
	tests := createAllBenchmarkStores()

	for _, test := range tests {
		b.Run(test.name, func(b *testing.B) {
			test.store.Put(key, value)

			if test.store.Len() != 1 {
				b.Errorf("s.Len() = %d, want %d", test.store.Len(), 1)
			}

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				test.store.Get(key)
			}
		})
	}
}

func BenchmarkStorePutTheSameConcurrently(b *testing.B) {
	key, value := "my-key", "my-value"
	tests := createAllBenchmarkStores()

	for _, test := range tests {
		if !test.store.IsThreadSafe() {
			b.Skipf("skipping test because the store is not thread-safe and we will get %q", "fatal error: concurrent map writes")
		}

		b.Run(test.name, func(b *testing.B) {
			b.ResetTimer()

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

func BenchmarkStoreGetTheSameConcurrently(b *testing.B) {
	key, value := "my-key", "my-value"
	tests := createAllBenchmarkStores()

	for _, test := range tests {
		if !test.store.IsThreadSafe() {
			b.Skipf("skipping test because the store is not thread-safe and we will get %q", "fatal error: concurrent map writes")
		}

		b.Run(test.name, func(b *testing.B) {
			test.store.Put(key, value)

			if test.store.Len() != 1 {
				b.Errorf("s.Len() = %d, want %d", test.store.Len(), 1)
			}

			b.ResetTimer()

			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					test.store.Get(key)
				}
			})

			if test.store.Len() != 1 {
				b.Errorf("s.Len() = %d, want %d", test.store.Len(), 1)
			}
		})
	}
}
