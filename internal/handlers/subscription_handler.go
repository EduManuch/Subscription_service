package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	serv "sub_service/internal/service"
)

type SubscriptionHandler struct {
	service *serv.SubscriptionService
}

func NewSubscriptionHandler(service *serv.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{service: service}
}

func (h *SubscriptionHandler) Create(w http.ResponseWriter, r *http.Request) {
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(response)
}

func (h *SubscriptionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
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

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

func (h *SubscriptionHandler) List(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	var userID *string
	if v := query.Get("user_id"); v != "" {
		userID = &v
	}

	var serviceName *string
	if v := query.Get("service_name"); v != "" {
		serviceName = &v
	}

	limit := 10
	if v := query.Get("limit"); v != "" {
		parsed, err := strconv.Atoi(v)
		if err != nil {
			http.Error(w, "invalid limit", http.StatusBadRequest)
			return
		}
		limit = parsed
	}

	offset := 0
	if v := query.Get("offset"); v != "" {
		parsed, err := strconv.Atoi(v)
		if err != nil {
			http.Error(w, "invalid offset", http.StatusBadRequest)
			return
		}
		offset = parsed
	}

	subscriptions, err := h.service.List(r.Context(), serv.ListSubscriptionsInput{
		UserID:      userID,
		ServiceName: serviceName,
		Limit:       limit,
		Offset:      offset,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := make([]map[string]any, 0, len(subscriptions))
	for _, s := range subscriptions {
		sub := map[string]any{
			"id":           s.ID,
			"service_name": s.ServiceName,
			"price":        s.Price,
			"user_id":      s.UserID,
			"start_date":   s.StartDate.Format("01-2006"),
			"created_at":   s.CreatedAt,
			"updated_at":   s.UpdatedAt,
		}

		if s.EndDate != nil {
			sub["end_date"] = s.EndDate.Format("01-2006")
		}

		response = append(response, sub)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

func (h *SubscriptionHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var input serv.UpdateSubscriptionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	sub, err := h.service.Update(r.Context(), id, input)
	if err != nil {
		switch {
		case errors.Is(err, serv.ErrSubNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	response := map[string]any{
		"id":           sub.ID,
		"service_name": sub.ServiceName,
		"price":        sub.Price,
		"user_id":      sub.UserID,
		"start_date":   sub.StartDate.Format("01-2006"),
		"updated_at":   sub.UpdatedAt,
	}

	if sub.EndDate != nil {
		response["end_date"] = sub.EndDate.Format("01-2006")
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}
