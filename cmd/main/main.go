package main

import (
	"Infotecs/internal/handlers"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/api/send", handlers.HandlerSend)
	http.HandleFunc("/api/transactions", handlers.HandlerGetLast)
	http.HandleFunc("/api/wallet/", handlers.HandlerGetBalance)

	log.Println("Server started on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
