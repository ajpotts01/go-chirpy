package main

import (
	"log"
	"net/http"
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

	mux := http.NewServeMux()

	mux.Handle("/app", cfg.metrics(http.StripPrefix("/app", http.FileServer(http.Dir(dirRoot)))))
	mux.HandleFunc(healthEndpoint, ready)
	mux.HandleFunc(metricsEndpoint, cfg.hits)

	corsHandler := cors(mux)

	// Can just do http.ListenAndServe but it may be useful to keep the server object around
	server := &http.Server{
		Addr:    ":" + port,
		Handler: corsHandler,
	}

	log.Printf("Now serving on port: %v", port)
	log.Fatal(server.ListenAndServe())
}
