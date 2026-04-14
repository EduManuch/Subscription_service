package handlers

import (
	"encoding/json"
	"net/http"
	serv "sub_service/internal/service"
)

type SubscriptionHandler struct {
	service *serv.SubscriptionService
}

func NewSubscriptionHandler(service *serv.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{service: service}
}

func (h *SubscriptionHandler) Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var input serv.CreationSubscriptionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	subscription, err := h.service.Create(r.Context(), input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := map[string]any{
		"id":           subscription.ID,
		"service_name": subscription.ServiceName,
		"price":        subscription.Price,
		"user_id":      subscription.UserID,
		"start_date":   subscription.StartDate.Format("01-2006"),
		"created_at":   subscription.CreatedAt,
		"updated_at":   subscription.UpdatedAt,
	}

	if subscription.EndDate != nil {
		response["end_date"] = subscription.EndDate.Format("01-2006")
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(response)
}
