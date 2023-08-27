package main

import (
	"log"
	"net/http"
)

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "*")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	const port = "8080"
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(".")))
	corsHandler := cors(mux)

	// Can just do http.ListenAndServe but it may be useful to keep the server object around
	server := &http.Server{
		Addr:    ":" + port,
		Handler: corsHandler,
	}

	log.Printf("Now serving on port: %v", port)
	log.Fatal(server.ListenAndServe())
}
