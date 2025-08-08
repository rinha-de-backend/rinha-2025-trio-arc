package main

import (
	"log"
	"net/http"

	"github.com/rinha-de-backend/rinha-2025-trio-arc/internal/handlers"
)

func main() {
	http.HandleFunc("/payments", handlers.HandlePayments)
	http.HandleFunc("/payments-summary", handlers.HandlePaymentsSummary)

	port := ":8080"
	log.Printf("Server listening on %s", port)

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
