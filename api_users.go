package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"
)

type userParams struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userAuthParams struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// No token or passwords returned
type userReturn struct {
	Id    int    `json:"id"`
	Email string `json:"email"`
}

type userAuthReturn struct {
	Id           int    `json:"id"`
	Email        string `json:"string"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

type userUpdateParams struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshTokenReturn struct {
	Token string `json:"token"`
}

// POST /api/users
func (config *apiConfig) createUser(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
	params := userParams{}
	err := decoder.Decode(&params)

	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	newUser, err := config.DbConn.CreateUser(params.Email, params.Password)

	if err != nil {
		log.Printf("Error creating new User: %v", err)
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	validResponse(w, http.StatusCreated, userReturn{
		Id:    newUser.Id,
		Email: newUser.Email,
	})
	return
}

// GET /api/users/{id}
func (config *apiConfig) readUser(w http.ResponseWriter, r *http.Request) {
	providedId := chi.URLParam(r, "id")

	id, err := strconv.Atoi(providedId)

	if err != nil {
		id = 0
	}

	user, err := config.DbConn.ReadUser(id)

	if err != nil {
		if err == os.ErrNotExist {
			errorResponse(w, http.StatusNotFound, err.Error())
			return
		}

		log.Printf("Error creating database connection: %v", err)
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	validResponse(w, http.StatusOK, user)
	return
}

// POST /api/users/login
func (config *apiConfig) authUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	params := userAuthParams{}
	err := decoder.Decode(&params)

	if err != nil {
		log.Printf("%v error getting parameters: %v\n", http.StatusInternalServerError, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	authUser, err := config.DbConn.AuthUser(params.Email, params.Password)

	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			log.Printf("%v error authorising user: %v\n", http.StatusUnauthorized, err)
			errorResponse(w, http.StatusUnauthorized, err.Error())
			return
		}

		log.Printf("%v error authorising user: %v\n", http.StatusNotFound, err)
		errorResponse(w, http.StatusNotFound, err.Error())
		return
	}

	accessToken, err := getJwt("chirpy-access", getAccessTokenExpiry(), fmt.Sprint(authUser.Id))

	if err != nil {
		log.Printf("%v error getting access token: %v\n", http.StatusInternalServerError, err)
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	refreshToken, err := getJwt("chirpy-refresh", getRefreshTokenExpiry(), fmt.Sprint(authUser.Id))

	if err != nil {
		log.Printf("%v error getting refresh token: %v\n", http.StatusInternalServerError, err)
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("Returning valid authorized user.")
	validResponse(w, http.StatusOK, userAuthReturn{
		Email:        authUser.Email,
		Id:           authUser.Id,
		Token:        accessToken,
		RefreshToken: refreshToken,
	})
	return
}

// PUT /api/users
// TODO: This function is far too long
func (config *apiConfig) updateUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	params := userUpdateParams{}
	err := decoder.Decode(&params)
	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		log.Printf("%v error getting parameters: %v\n", http.StatusInternalServerError, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	suppliedToken, err := getSuppliedToken(r)

	if err != nil {
		errorResponse(w, http.StatusUnauthorized, "Bad authorization header")
		return
	}

	claims, err := checkToken(suppliedToken, "chirpy-access")

	if err != nil {
		errorResponse(w, http.StatusUnauthorized, err.Error())
		return
	}

	id, err := strconv.Atoi(claims.Subject)

	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "Failed to parse user ID")
		return
	}

	updatedUser, err := config.DbConn.UpdateUser(id, params.Email, params.Password)

	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	validResponse(w, http.StatusOK, userReturn{
		Id:    updatedUser.Id,
		Email: updatedUser.Email,
	})
	return

}
