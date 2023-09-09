package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

type chirpReturn struct {
	Body     string `json:"body"`
	AuthorId int    `json:"author_id"`
	Id       int    `json:"id"`
}

type chirpParams struct {
	Body string `json:"body"`
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

func (config *apiConfig) readChirp(w http.ResponseWriter, r *http.Request) {
	providedId := chi.URLParam(r, "id")
	id, err := strconv.Atoi(providedId)

	if err != nil {
		id = 0
	}

	if id == 0 {
		providedAuthorId := r.URL.Query().Get("author_id")
		authorId, err := strconv.Atoi(providedAuthorId)

		if err != nil {
			authorId = 0
		}

		sort := "asc"
		providedSort := r.URL.Query().Get("sort")

		if providedSort != "" {
			sort = providedSort
		}

		chirps, err := config.DbConn.ReadChirps(authorId, sort)

		if err != nil {
			log.Printf("Error creating database connection: %v", err.Error())
			errorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		fmt.Printf("Chirps: %v\n", chirps)

		validResponse(w, http.StatusOK, chirps)
		return

	} else {
		chirp, err := config.DbConn.ReadSingleChirp(id)

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

func (config *apiConfig) createChirp(w http.ResponseWriter, r *http.Request) {
	suppliedToken, err := getAuthHeaderItem(r, "Bearer")
	if err != nil {
		errorResponse(w, http.StatusUnauthorized, "Bad authorization header")
		return
	}

	claims, err := checkToken(suppliedToken, "chirpy-access")
	if err != nil {
		errorResponse(w, http.StatusUnauthorized, "Bad token")
		return
	}

	authorId, err := strconv.Atoi(claims.Subject)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "Could not determine author ID")
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := chirpParams{}
	err = decoder.Decode(&params)

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

	newChirp, err := config.DbConn.CreateChirp(cleanedBody, authorId)

	if err != nil {
		log.Printf("Error creating new Chirp: %v", err.Error())
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	validResponse(w, http.StatusCreated, chirpReturn{
		Id:       newChirp.Id,
		Body:     newChirp.Body,
		AuthorId: authorId,
	})
	return
}

func (config *apiConfig) deleteChirp(w http.ResponseWriter, r *http.Request) {
	suppliedToken, err := getAuthHeaderItem(r, "Bearer")
	if err != nil {
		errorResponse(w, http.StatusUnauthorized, "Bad authorization header")
		return
	}

	claims, err := checkToken(suppliedToken, "chirpy-access")
	if err != nil {
		errorResponse(w, http.StatusUnauthorized, "Bad token")
		return
	}

	userId, err := strconv.Atoi(claims.Subject)

	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	providedChirpId := chi.URLParam(r, "id")
	chirpId, err := strconv.Atoi(providedChirpId)

	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	chirp, err := config.DbConn.ReadSingleChirp(chirpId)

	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	if chirp.AuthorId != userId {
		errorResponse(w, http.StatusForbidden, "Cannot delete a chirp you didn't post")
		return
	}

	err = config.DbConn.DeleteSingleChirp(chirpId)

	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	return
}
