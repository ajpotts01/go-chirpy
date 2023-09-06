package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/ajpotts01/go-chirpy/internal/database"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type userParams struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userAuthParams struct {
	Email            string `json:"email"`
	Password         string `json:"password"`
	ExpiresInSeconds int    `json:"expires_in_seconds"`
}

// No token or passwords returned
type userReturn struct {
	Id    int    `json:"id"`
	Email string `json:"email"`
}

type userAuthReturn struct {
	Id    int    `json:"id"`
	Email string `json:"string"`
	Token string `json:"token"`
}

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

	fmt.Println("Database connection open...")

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

	token, err := getJwt(params.ExpiresInSeconds, authUser)

	if err != nil {
		log.Printf("%v error getting token: %v\n", http.StatusInternalServerError, err)
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("Returning valid authorized user.")
	validResponse(w, http.StatusOK, userAuthReturn{
		Email: authUser.Email,
		Id:    authUser.Id,
		Token: token,
	})
	return
}

func getJwt(expiry int, user database.User) (string, error) {
	claims := jwt.RegisteredClaims{
		Issuer: "chirpy",
		IssuedAt: &jwt.NumericDate{
			time.Now().UTC(),
		},
		ExpiresAt: &jwt.NumericDate{
			time.Now().UTC().Add(time.Duration(expiry * int(time.Second))),
		},
		Subject: fmt.Sprint(user.Id),
	}
	log.Println("Claims set up")

	newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	log.Println("Token created")

	fmt.Println(os.Getenv("JWT_SECRET"))
	tokenStr, err := newToken.SignedString([]byte(os.Getenv("JWT_SECRET")))

	if err != nil {
		fmt.Printf("Failed to sign token: %v\n", err)
		return "", err
	}

	return tokenStr, nil
}
