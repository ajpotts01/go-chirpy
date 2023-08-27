package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {
	const port = "8080"
	const dirRoot = "."
	const healthEndpoint = "/healthz"
	const metricsEndpoint = "/metrics"
	const appEndpoint = "/app"

	cfg := apiConfig{
		serverHits: 0,
	}

	router := chi.NewRouter()
	fsHandler := cfg.metrics(http.StripPrefix(appEndpoint, http.FileServer(http.Dir(dirRoot))))

	// Done a bit differently to the boot.dev example
	// They just use router in the same way as mux
	// e.g. corsHandler := cors(mux) => corsHandler := cors(router)
	// But router.Use works just as well

	router.Use(cors)
	router.Handle("/app", fsHandler)
	router.Handle("/app/*", fsHandler)
	router.Get(healthEndpoint, ready)
	router.Get(metricsEndpoint, cfg.hits)

	// Can just do http.ListenAndServe but it may be useful to keep the server object around
	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	log.Printf("Now serving on port: %v", port)
	log.Fatal(server.ListenAndServe())
}
