package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/rinha-de-backend/rinha-2025-trio-arc/internal/models"
)

func HandlePayments(w http.ResponseWriter, r *http.Request) {
	var request models.PaymentRequest

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if request.CorrelationID == "" || request.Amount <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// payment := models.NewPayment(request.CorrelationID, request.Amount)
	// serialized, err := json.Marshal(payment)
	// if err != nil {
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	return
	// }

	// push serialized message to Redis

	w.WriteHeader(http.StatusOK)
}
