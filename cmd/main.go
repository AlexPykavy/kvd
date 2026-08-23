package main

import (
	"flag"
	"kvd/api/middleware"
	v1api "kvd/api/v1"
	v0 "kvd/internal/store/v0"
	v0store "kvd/internal/store/v0"
	"log/slog"
	"os"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"

	"log"
	"net/http"
)

func main() {
	capacity := flag.Int("capacity", 0, "Store capacity")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	options := []v0.StoreOption{}
	if *capacity > 0 {
		logger.Info("Initializing the store with preset capacity", slog.Int("capacity", *capacity))

		options = append(options, v0.WithCapacity(*capacity))
	}

	s := v0store.NewStore(options...)
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
		log.Fatal(err)
	}
}
