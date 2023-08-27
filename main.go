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
	const validateEndpoint = "/validate_chirp"

	cfg := apiConfig{
		serverHits: 0,
	}

	appRouter := chi.NewRouter()
	fsHandler := cfg.metrics(http.StripPrefix(appEndpoint, http.FileServer(http.Dir(dirRoot))))

	apiRouter := chi.NewRouter()
	apiRouter.Get(healthEndpoint, ready)
	apiRouter.Post(validateEndpoint, validateChirp)

	adminRouter := chi.NewRouter()
	adminRouter.Get(metricsEndpoint, cfg.hits)

	// Done a bit differently to the boot.dev example
	// They just use router in the same way as mux
	// e.g. corsHandler := cors(mux) => corsHandler := cors(router)
	// But router.Use works just as well
	appRouter.Use(cors)
	appRouter.Handle("/app", fsHandler)
	appRouter.Handle("/app/*", fsHandler)
	appRouter.Mount("/api", apiRouter)
	appRouter.Mount("/admin", adminRouter)

	// Can just do http.ListenAndServe but it may be useful to keep the server object around
	server := &http.Server{
		Addr:    ":" + port,
		Handler: appRouter,
	}

	log.Printf("Now serving on port: %v", port)
	log.Fatal(server.ListenAndServe())
}
