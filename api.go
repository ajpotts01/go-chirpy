package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/ajpotts01/go-chirpy/internal/database"
	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"
)

type apiConfig struct {
	serverHits int
	jwtSecret  string
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

func readChirp(w http.ResponseWriter, r *http.Request) {
	database, err := database.NewDatabase("database.json")

	if err != nil {
		log.Printf("Error creating database connection: %v", err.Error())
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	fmt.Println("Database connection open...")

	providedId := chi.URLParam(r, "id")

	id, err := strconv.Atoi(providedId)

	if err != nil {
		id = 0
	}

	if id == 0 {
		chirps, err := database.ReadChirps()

		if err != nil {
			log.Printf("Error creating database connection: %v", err.Error())
			errorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		fmt.Printf("Chirps: %v\n", chirps)

		validResponse(w, http.StatusOK, chirps)
		return

	} else {
		chirp, err := database.ReadSingleChirp(id)

		if err != nil {
			if err == os.ErrNotExist {
				errorResponse(w, http.StatusNotFound, err.Error())
				return
			}

			log.Printf("Error creating database connection: %v", err.Error())
			errorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		validResponse(w, http.StatusOK, chirp)
		return
	}
}

func createChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
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

	if len(cleanedBody) > 140 {
		errorResponse(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	database, err := database.NewDatabase("database.json")

	if err != nil {
		log.Printf("Error creating database connection: %v", err.Error())
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	newChirp, err := database.CreateChirp(cleanedBody)

	if err != nil {
		log.Printf("Error creating new Chirp: %v", err.Error())
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	validResponse(w, http.StatusCreated, newChirp)
	return
}

func createUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
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

	database, err := database.NewDatabase("database.json")

	if err != nil {
		log.Printf("Error creating database connection: %v", err.Error())
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	newUser, err := database.CreateUser(params.Email, params.Password)

	if err != nil {
		log.Printf("Error creating new User: %v", err.Error())
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	newUser.Password = nil
	validResponse(w, http.StatusCreated, newUser)
	return
}

func readUser(w http.ResponseWriter, r *http.Request) {
	database, err := database.NewDatabase("database.json")

	if err != nil {
		log.Printf("Error creating database connection: %v", err.Error())
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	fmt.Println("Database connection open...")

	providedId := chi.URLParam(r, "id")

	id, err := strconv.Atoi(providedId)

	if err != nil {
		id = 0
	}

	user, err := database.ReadUser(id)

	user.Password = nil // Will this make the password not return at all?

	if err != nil {
		if err == os.ErrNotExist {
			errorResponse(w, http.StatusNotFound, err.Error())
			return
		}

		log.Printf("Error creating database connection: %v", err.Error())
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	validResponse(w, http.StatusOK, user)
	return
}

func authUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
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

	database, err := database.NewDatabase("database.json")

	if err != nil {
		log.Printf("Error creating database connection: %v", err.Error())
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	fmt.Println("Database connection open...")

	authUser, err := database.AuthUser(params.Email, params.Password)

	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			errorResponse(w, http.StatusUnauthorized, err.Error())
			return
		}

		errorResponse(w, http.StatusNotFound, err.Error())
		return
	}

	validResponse(w, http.StatusOK, authUser)
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
