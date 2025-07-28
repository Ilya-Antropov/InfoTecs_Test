package handlers

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"

	"Infotecs/internal/database"
	"Infotecs/internal/models"
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
			http.Error(w, "Недостаточно средств", http.StatusPaymentRequired)
		default:
			http.Error(w, "Неудачная транзакция", http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *Handlers) HandlerGetBalance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	address := vars["address"]

	balance, err := h.db.GetWalletBalance(ctx, address)
	if err != nil {
		if errors.Is(err, database.ErrWalletNotFound) {
			http.Error(w, "Кошелек не найден", http.StatusNotFound)
		} else {
			http.Error(w, "Не удалось получить баланс", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]float64{"balance": balance})
}

func (h *Handlers) HandlerGetLast(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	countStr := vars["count"]

	count, err := strconv.Atoi(countStr)
	if err != nil || count < 1 {
		http.Error(w, "Неправильный параметр count", http.StatusBadRequest)
		return
	}

	transactions, err := h.db.GetTransactions(ctx, count)
	if err != nil {
		http.Error(w, "Не удалось получить транзакции", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transactions)
}
