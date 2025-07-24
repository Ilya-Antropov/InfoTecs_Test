package handlers

import (
	"Infotecs/internal/database"
	"Infotecs/internal/models"
	"encoding/json"
	"errors"
	"net/http"
)

type Handlers struct {
	db *database.DB
}

func NewHandlers(db *database.DB) *Handlers {
	return &Handlers{db: db}
}

func (h *Handlers) HandlerSend(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var request models.SendRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Недопустимый JSON", http.StatusBadRequest)
		return
	}

	if err := h.db.SendMoney(ctx, request.From, request.To, request.Amount); err != nil {
		switch {
		case errors.Is(err, database.ErrWalletNotFound):
			http.Error(w, "Кошелек отправителя не найден", http.StatusNotFound)
		case errors.Is(err, database.ErrRecipientNotFound):
			http.Error(w, "Кошелек получателя не найден", http.StatusNotFound)
		case errors.Is(err, database.ErrInsufficientBalance):
			http.Error(w, "Недостаточно средств", http.StatusBadRequest)
		default:
			http.Error(w, "Неудачная транзакция", http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func HandlerGetBalance(w http.ResponseWriter, r *http.Request) {

}

func HandlerGetLast(w http.ResponseWriter, r *http.Request) {

}
