package main

import (
	"context"
	"log"
	"net/http"

	"Infotecs/internal/database"
	"Infotecs/internal/handlers"

	"github.com/gorilla/mux"
)

func main() {
	connStr := "postgres://payment_users:payment_users@localhost/payment_system?sslmode=disable"
	db, err := database.InitDB(connStr)
	if err != nil {
		log.Fatal("База данных не запущена", err)
	}
	defer db.Close()

	if err := db.Initialize(context.Background()); err != nil {
		log.Fatal("Сбой в базе данных", err)
	}

	router := mux.NewRouter()
	h := handlers.NewHandlers(db)

	router.HandleFunc("/api/send", h.HandlerSend).Methods("POST")
	router.HandleFunc("/api/transactions/count/{count}", h.HandlerGetLast).Methods("GET")
	router.HandleFunc("/api/wallet/{address}/balance", h.HandlerGetBalance).Methods("GET")

	log.Println("Запуск сервера на порте:8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
