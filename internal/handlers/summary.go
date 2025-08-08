package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/rinha-de-backend/rinha-2025-trio-arc/internal/models"
)

func HandlePaymentsSummary(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	request, err := models.NewSummaryRequest(from, to)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	filter := &models.SummaryFilter{
		From: request.From,
		To:   request.To,
	}

	summary, err := models.GetStats(filter)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(summary)
}
