package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

type webhookData struct {
	UserId int `json:"user_id"`
}

type webhookEvent struct {
	Event string      `json:"event"`
	Data  webhookData `json:"data"`
}

func (config *apiConfig) upgradeUser(w http.ResponseWriter, r *http.Request) {
	apiKey, err := getAuthHeaderItem(r, "ApiKey")

	if err != nil || apiKey != os.Getenv("POLKA_KEY") {
		log.Printf("Error - bad API key")
		errorResponse(w, http.StatusUnauthorized, err.Error())
		return
	}

	log.Printf("Received data: %v", r.Body)
	decoder := json.NewDecoder(r.Body)
	event := webhookEvent{}
	err = decoder.Decode(&event)

	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("Received data on upgrade endpoint:\n")
	log.Printf("%s\n", event.Event)
	log.Printf("%d\n", event.Data.UserId)
	if event.Event != "user.upgraded" {
		log.Printf("User was not upgraded - event was %v", event.Event)
		w.WriteHeader(http.StatusOK)
		return
	}

	err = config.DbConn.UpgradeUser(event.Data.UserId)
	if err != nil {
		log.Printf("Error upgrading user: %s", err)
		errorResponse(w, http.StatusNotFound, err.Error())
		return
	}

	log.Printf("User was upgraded - event was %v", event.Event)
	w.WriteHeader(http.StatusOK)
	return
}
