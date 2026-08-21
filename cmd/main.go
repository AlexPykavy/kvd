package main

import (
	v1api "kvd/api/v1"
	v0store "kvd/internal/store/v0"

	httpSwagger "github.com/swaggo/http-swagger"

	"log"
	"net/http"
)

func main() {
	s := v0store.NewStore()
	v1Handler := v1api.NewStoreHandler(s)

	v1 := http.NewServeMux()
	v1.HandleFunc("PUT /keys/{key}", v1Handler.Put)
	v1.HandleFunc("GET /keys/{key}", v1Handler.Get)
	v1.HandleFunc("DELETE /keys/{key}", v1Handler.Delete)
	v1.HandleFunc("GET /count", v1Handler.Count)
	v1.HandleFunc("GET /swagger/spec", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./docs/v1/swagger.json")
	})

	mux := http.NewServeMux()
	mux.Handle("/v1/", http.StripPrefix("/v1", v1))
	mux.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.UIConfig(map[string]string{
			"urls": `[
				{"url": "/v1/swagger/spec", "name": "API v1"},
			]`,
		}),
	))

	log.Println("server listening on :8080")

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
