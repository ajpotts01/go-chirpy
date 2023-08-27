package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

type apiConfig struct {
	serverHits int
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

func profanityFilter(body string) string {
	profanity := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}

	censor := "****"

	words := strings.Split(body, " ")

	for idx, word := range words {
		if _, ok := profanity[strings.ToLower(word)]; ok {
			words[idx] = censor
		}
	}

	result := strings.Join(words, " ")
	return result
}

func validateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	type successResponse struct {
		CleanedBody string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)

	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	cleanedBody := profanityFilter(params.Body)

	log.Printf("Received chirp with length of %v\n", len(params.Body))

	if len(params.Body) > 140 {
		response := errorResponse{
			Err: "Chirp is too long",
		}

		data, err := json.Marshal(response)

		if err != nil {
			log.Println(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Header needs to be done before data is written
		// or the header is automatically 200
		w.WriteHeader(http.StatusBadRequest)
		w.Write(data)
		return
	}

	response := successResponse{
		CleanedBody: cleanedBody,
	}

	data, err := json.Marshal(response)

	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
	return
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
