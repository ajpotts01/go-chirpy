package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/ajpotts01/go-chirpy/internal/database"
)

type apiConfig struct {
	serverHits int
	jwtSecret  string
	DbConn     *database.Database
}

func (cfg *apiConfig) metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.serverHits += 1
		log.Printf(fmt.Sprintf("Hit logged. Total hits: %v", cfg.serverHits))
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) hits(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`
	<html>
	
	<body>
		<h1>Welcome, Chirpy Admin</h1>
		<p>Chirpy has been visited %d times!</p>
	</body>
	
	</html>
		`, cfg.serverHits)))
}

func validResponse(w http.ResponseWriter, code int, obj interface{}) {
	w.Header().Set("Content-Type", "application/json")

	resp, err := json.Marshal(obj)

	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(code)
	w.Write(resp)
}

func errorResponse(w http.ResponseWriter, code int, errorMsg string) {
	type errorResponse struct {
		Err string `json:"error"`
	}

	errorObj := errorResponse{
		Err: errorMsg,
	}

	validResponse(w, code, errorObj)
}
