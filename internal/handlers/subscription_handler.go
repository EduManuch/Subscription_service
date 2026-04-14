package handlers

import (
	"encoding/json"
	"errors"
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

	sub, err := h.service.Create(r.Context(), input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := map[string]any{
		"id":           sub.ID,
		"service_name": sub.ServiceName,
		"price":        sub.Price,
		"user_id":      sub.UserID,
		"start_date":   sub.StartDate.Format("01-2006"),
		"created_at":   sub.CreatedAt,
		"updated_at":   sub.UpdatedAt,
	}

	if sub.EndDate != nil {
		response["end_date"] = sub.EndDate.Format("01-2006")
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(response)
}

func (h *SubscriptionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	id := r.PathValue("id")

	sub, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, serv.ErrSubNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		case errors.Is(err, serv.ErrInvalidSubID):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	response := map[string]any{
		"id":           sub.ID,
		"service_name": sub.ServiceName,
		"price":        sub.Price,
		"user_id":      sub.UserID,
		"start_date":   sub.StartDate.Format("01-2006"),
		"created_at":   sub.CreatedAt,
		"updated_at":   sub.UpdatedAt,
	}

	if sub.EndDate != nil {
		response["end_date"] = sub.EndDate.Format("01-2006")
	}

	_ = json.NewEncoder(w).Encode(response)
}
