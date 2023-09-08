package main

import (
	"log"
	"net/http"
	"os"

	"github.com/ajpotts01/go-chirpy/internal/database"
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

func main() {
	const port = "8080"
	const dirRoot = "."
	const healthEndpoint = "/healthz"
	const metricsEndpoint = "/metrics"
	const appEndpoint = "/app"
	const chirpEndpoint = "/chirps"
	const singleChirpEndpoint = "/chirps/{id}"
	const userEndpoint = "/users"
	const loginEndpoint = "/login"

	godotenv.Load()

	dbConn, err := database.NewDatabase("database.json")

	if err != nil {
		log.Fatal(err)
	}

	config := apiConfig{
		serverHits: 0,
		jwtSecret:  os.Getenv("JWT_SECRET"),
		DbConn:     dbConn,
	}

	appRouter := chi.NewRouter()
	fsHandler := config.metrics(http.StripPrefix(appEndpoint, http.FileServer(http.Dir(dirRoot))))

	apiRouter := chi.NewRouter()
	apiRouter.Get(healthEndpoint, ready)
	apiRouter.Get(chirpEndpoint, config.readChirp)
	apiRouter.Get(singleChirpEndpoint, config.readChirp)
	apiRouter.Post(chirpEndpoint, config.createChirp)
	apiRouter.Post(userEndpoint, config.createUser)
	apiRouter.Get(userEndpoint, config.readUser)
	apiRouter.Put(userEndpoint, config.updateUser)
	apiRouter.Post(loginEndpoint, config.authUser)

	adminRouter := chi.NewRouter()
	adminRouter.Get(metricsEndpoint, config.hits)

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
