package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// POST /api/refresh
func (config *apiConfig) refreshToken(w http.ResponseWriter, r *http.Request) {
	// No body, just check headers
	authHeader := r.Header.Get("Authorization")

	if authHeader == "" {
		log.Printf("Unauthorized")
		errorResponse(w, http.StatusBadRequest, "Must provide authorization header")
		return
	}

	suppliedToken := strings.Replace(authHeader, "Bearer ", "", 1)

	// TODO: If revoked
	revoked, err := config.DbConn.IsTokenRevoked(suppliedToken)
	if revoked {
		errorResponse(w, http.StatusUnauthorized, "This refresh token has been revoked")
		return
	}

	token, err := jwt.ParseWithClaims(suppliedToken, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		log.Printf("Error parsing token")
		errorResponse(w, http.StatusUnauthorized, err.Error())
		return
	}

	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		if claims.Issuer != "chirpy-refresh" {
			errorResponse(w, http.StatusUnauthorized, "Only refresh tokens are accepted")
		}

		token, err := getJwt("chirpy-refresh", getRefreshTokenExpiry(), claims.Subject)

		if err != nil {
			log.Printf("%v error getting token: %v\n", http.StatusInternalServerError, err)
			errorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		log.Printf("Returning valid refresh token.")
		validResponse(w, http.StatusOK, refreshTokenReturn{
			Token: token,
		})
		return

	} else {
		errorResponse(w, http.StatusUnauthorized, "Bad token")
		return
	}

}

func (config *apiConfig) revokeToken(w http.ResponseWriter, r *http.Request) {
	// No body, just check headers
	authHeader := r.Header.Get("Authorization")

	if authHeader == "" {
		log.Printf("Unauthorized")
		errorResponse(w, http.StatusBadRequest, "Must provide authorization header")
		return
	}

	suppliedToken := strings.Replace(authHeader, "Bearer ", "", 1)

	err := config.DbConn.RevokeToken(suppliedToken)

	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "Error revoking token")
		return
	}

	w.WriteHeader(http.StatusOK)
	return
}

func getAccessTokenExpiry() time.Time {
	return time.Now().UTC().Add(time.Duration(1 * int(time.Hour)))
}

func getRefreshTokenExpiry() time.Time {
	return time.Now().UTC().Add(time.Duration(60 * 24 * int(time.Hour)))
}

func getJwt(issuer string, expiresAt time.Time, subject string) (string, error) {
	claims := jwt.RegisteredClaims{
		Issuer: issuer,
		IssuedAt: &jwt.NumericDate{
			time.Now().UTC(),
		},
		ExpiresAt: &jwt.NumericDate{
			expiresAt,
		},
		Subject: subject,
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
