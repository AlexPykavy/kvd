package main

import (
	"flag"
	"kvd/api/middleware"
	v1api "kvd/api/v1"
	v2store "kvd/internal/store/v2"
	"log/slog"
	"os"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"

	"net/http"
)

const (
	mutexTypeStub        = "stub"
	mutexTypeSyncMutex   = "sync.Mutex"
	mutexTypeSyncRWMutex = "sync.RWMutex"
)

func main() {
	capacity := flag.Uint64("capacity", 0, "Store capacity")
	shards := flag.Uint64("shards", 1, "Numer of locking shards")
	mutexType := flag.String("mutex-type", mutexTypeStub, "Mutex type (stub|sync.Mutex|sync.RWMutex|). Default is stub")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	options := []v2store.MyHashTableOption{}
	if *capacity > 0 {
		logger.Info("Initializing the store with preset capacity", slog.Uint64("capacity", *capacity))

		options = append(options, v2store.WithCapacity(*capacity))
	}

	switch *mutexType {
	case mutexTypeStub:
	case mutexTypeSyncMutex:
		options = append(options, v2store.WithMutex(*shards))
	case mutexTypeSyncRWMutex:
		options = append(options, v2store.WithRWMutex(*shards))
	default:
		logger.Error("Unsupported --mutex-type", "type", *mutexType)
		os.Exit(1)
	}

	logger.Info("Initializing the store with a mutex", "type", *mutexType, "shards", shards)

	s := v2store.NewMyHashTable(options...)
	v1Handler := v1api.NewStoreHandler(s)

	v1 := http.NewServeMux()
	v1.HandleFunc("PUT /keys/{key}", v1Handler.Put)
	v1.HandleFunc("GET /keys/{key}", v1Handler.Get)
	v1.HandleFunc("DELETE /keys/{key}", v1Handler.Delete)
	v1.HandleFunc("GET /count", v1Handler.Count)

	mux := http.NewServeMux()
	mux.Handle("/v1/", http.StripPrefix("/v1", middleware.InstrumentMiddleware(logger)(v1)))
	mux.HandleFunc("GET /v1/swagger/spec", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./docs/v1/swagger.json")
	})

	mux.Handle("/metrics", promhttp.Handler())
	mux.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.UIConfig(map[string]string{
			"urls": `[
				{"url": "/v1/swagger/spec", "name": "API v1"},
			]`,
		}),
	))

	logger.Info("server listening", slog.String("address", ":8080"))

	if err := http.ListenAndServe(":8080", mux); err != nil {
		slog.Error("Failed to start the server", "error", err)
		os.Exit(1)
	}
}
