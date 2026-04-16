package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"sub_service/internal/model"
	serv "sub_service/internal/service"
	"time"

	"github.com/google/uuid"
)

type SubscriptionHandler struct {
	service *serv.SubscriptionService
}

func NewSubscriptionHandler(service *serv.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{service: service}
}

// CRLResponse - Create, Read, List response
// Ответ для Create, GetByID и List
type CRLResponse struct {
	ID          uuid.UUID
	ServiceName string    `json:"service_name"`
	Price       int       `json:"price"`
	UserID      uuid.UUID `json:"user_id"`
	StartDate   string    `json:"start_date"`
	EndDate     string    `json:"end_date,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (h *SubscriptionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input serv.CreationSubscriptionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	sub, err := h.service.Create(r.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, serv.ErrServiceNameRequired) ||
			errors.Is(err, serv.ErrPriceLessThan0) ||
			errors.Is(err, serv.ErrInvalidUserID) ||
			errors.Is(err, serv.ErrInvalidStartDate) ||
			errors.Is(err, serv.ErrInvalidEndDate) ||
			errors.Is(err, serv.ErrEndDateGreaterStartDate):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	response := createResponse(sub)

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

	response := createResponse(sub)

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
		case errors.Is(err, serv.ErrInvalidSubID) ||
			errors.Is(err, serv.ErrServiceNameRequired) ||
			errors.Is(err, serv.ErrPriceLessThan0) ||
			errors.Is(err, serv.ErrInvalidUserID) ||
			errors.Is(err, serv.ErrInvalidStartDate) ||
			errors.Is(err, serv.ErrInvalidEndDate) ||
			errors.Is(err, serv.ErrEndDateGreaterStartDate):
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
		"updated_at":   sub.UpdatedAt,
	}

	if sub.EndDate != nil {
		response["end_date"] = sub.EndDate.Format("01-2006")
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

func (h *SubscriptionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	err := h.service.Delete(r.Context(), id)
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

	w.WriteHeader(http.StatusNoContent)
}

func (h *SubscriptionHandler) CalculateTotalPrice(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	from := query.Get("from")
	to := query.Get("to")

	if from == "" || to == "" {
		http.Error(w, "from and to dates are required", http.StatusBadRequest)
		return
	}

	var userID *string
	if v := query.Get("user_id"); v != "" {
		userID = &v
	}

	var serviceName *string
	if v := query.Get("service_name"); v != "" {
		serviceName = &v
	}

	total, err := h.service.CalculateTotalPrice(r.Context(), serv.CalculateTotalPriceInput{
		From:        from,
		To:          to,
		UserID:      userID,
		ServiceName: serviceName,
	})

	if err != nil {
		switch {
		case errors.Is(err, serv.ErrInvalidStartDate) ||
			errors.Is(err, serv.ErrInvalidEndDate) ||
			errors.Is(err, serv.ErrEndDateGreaterStartDate) ||
			errors.Is(err, serv.ErrInvalidUserID):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	response := map[string]any{
		"total": total,
		"from":  from,
		"to":    to,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

func createResponse(sub *model.Subscription) *CRLResponse {
	response := CRLResponse{
		ID:          sub.ID,
		ServiceName: sub.ServiceName,
		Price:       sub.Price,
		UserID:      sub.UserID,
		StartDate:   sub.StartDate.Format("01-2006"),
		CreatedAt:   sub.CreatedAt,
		UpdatedAt:   sub.UpdatedAt,
	}

	if sub.EndDate != nil {
		response.EndDate = sub.EndDate.Format("01-2006")
	}

	return &response
}
