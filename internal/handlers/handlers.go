package handlers

import (
	"Infotecs/internal/models"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

func HandlerSend(w http.ResponseWriter, r *http.Request) {
	var tx models.Transaction
	if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

}

func HandlerGetBalance(w http.ResponseWriter, r *http.Request) {

}

func HandlerGetLast(w http.ResponseWriter, r *http.Request) {

}
