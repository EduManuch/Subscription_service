package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"sub_service/internal/model"
	serv "sub_service/internal/service"
)

type SubscriptionHandler struct {
	service *serv.SubscriptionService
	logger  *slog.Logger
}

func NewSubscriptionHandler(service *serv.SubscriptionService, logger *slog.Logger) *SubscriptionHandler {
	return &SubscriptionHandler{
		service: service,
		logger:  logger,
	}
}

// Create godoc
// @Summary Create subscription
// @Description Create a new subscription
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param input body service.CreationSubscriptionInput true "Subscription payload"
// @Success 201 {object} handlers.CRLResponse
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 500 {object} handlers.ErrorResponse
// @Router /subscriptions [post]
func (h *SubscriptionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input serv.CreationSubscriptionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		h.logger.Debug("invalid request body", "error", err)
		return
	}

	sub, err := h.service.Create(r.Context(), input)
	if h.responseError(w, err) {
		return
	}

	response := createResponse(sub)

	h.writeJSON(w, http.StatusCreated, response)
	h.logger.Info("subscription created", "id", sub.ID)
}

// GetByID godoc
// @Summary Get subscription by ID
// @Tags subscriptions
// @Produce json
// @Param id path string true "Subscription ID"
// @Success 200 {object} handlers.CRLResponse
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 404 {object} handlers.ErrorResponse
// @Failure 500 {object} handlers.ErrorResponse
// @Router /subscriptions/{id} [get]
func (h *SubscriptionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	sub, err := h.service.GetByID(r.Context(), id)
	if h.responseError(w, err) {
		return
	}

	response := createResponse(sub)

	h.writeJSON(w, http.StatusOK, response)
}

// List godoc
// @Summary List subscriptions
// @Tags subscriptions
// @Produce json
// @Param user_id query string false "User ID"
// @Param service_name query string false "Service name"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {array} handlers.CRLResponse
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 500 {object} handlers.ErrorResponse
// @Router /subscriptions [get]
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
			h.writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid limit"})
			h.logger.Debug("invalid limit", "error", err)
			return
		}
		limit = parsed
	}

	offset := 0
	if v := query.Get("offset"); v != "" {
		parsed, err := strconv.Atoi(v)
		if err != nil {
			h.writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid offset"})
			h.logger.Debug("invalid offset", "error", err)
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

	if h.responseError(w, err) {
		return
	}

	response := make([]CRLResponse, 0, len(subscriptions))
	for _, s := range subscriptions {
		resp := createResponse(&s)
		response = append(response, resp)
	}

	h.writeJSON(w, http.StatusOK, response)
	h.logger.Info("subscriptions listed", "count", len(subscriptions))
}

// Update godoc
// @Summary Update subscription
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "Subscription ID"
// @Param input body service.UpdateSubscriptionInput true "Subscription payload"
// @Success 200 {object} handlers.CRLResponse
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 404 {object} handlers.ErrorResponse
// @Failure 500 {object} handlers.ErrorResponse
// @Router /subscriptions/{id} [put]
func (h *SubscriptionHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var input serv.UpdateSubscriptionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		h.logger.Debug("invalid request body", "error", err)
		return
	}

	sub, err := h.service.Update(r.Context(), id, input)
	if h.responseError(w, err) {
		return
	}

	response := createResponse(sub)

	h.writeJSON(w, http.StatusOK, response)
	h.logger.Info("subscription updated", "id", sub.ID)
}

// Delete godoc
// @Summary Delete subscription
// @Tags subscriptions
// @Produce json
// @Param id path string true "Subscription ID"
// @Success 204
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 404 {object} handlers.ErrorResponse
// @Failure 500 {object} handlers.ErrorResponse
// @Router /subscriptions/{id} [delete]
func (h *SubscriptionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	err := h.service.Delete(r.Context(), id)
	if h.responseError(w, err) {
		return
	}

	w.WriteHeader(http.StatusNoContent)
	h.logger.Info("subscription deleted", "id", id)
}

// CalculateTotalPrice godoc
// @Summary Calculate total subscription price
// @Tags subscriptions
// @Produce json
// @Param from query string true "Start period in MM-YYYY"
// @Param to query string true "End period in MM-YYYY"
// @Param user_id query string false "User ID"
// @Param service_name query string false "Service name"
// @Success 200 {object} handlers.TotalPriceResponse
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 500 {object} handlers.ErrorResponse
// @Router /subscriptions/total [get]
func (h *SubscriptionHandler) CalculateTotalPrice(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	from := query.Get("from")
	to := query.Get("to")

	if from == "" || to == "" {
		h.writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "from and to dates are required"})
		h.logger.Debug("missing required query params", "from", from, "to", to)
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

	if h.responseError(w, err) {
		return
	}

	response := TotalPriceResponse{
		Total: total,
		From:  from,
		To:    to,
	}

	h.writeJSON(w, http.StatusOK, response)
}

func createResponse(sub *model.Subscription) CRLResponse {
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

	return response
}

func (h *SubscriptionHandler) responseError(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	switch {
	case errors.Is(err, serv.ErrSubNotFound):
		h.writeJSON(w, http.StatusNotFound, ErrorResponse{Error: err.Error()})
	case errors.Is(err, serv.ErrInvalidSubID) ||
		errors.Is(err, serv.ErrServiceNameRequired) ||
		errors.Is(err, serv.ErrPriceLessThan0) ||
		errors.Is(err, serv.ErrInvalidUserID) ||
		errors.Is(err, serv.ErrInvalidStartDate) ||
		errors.Is(err, serv.ErrInvalidEndDate) ||
		errors.Is(err, serv.ErrEndDateGreaterStartDate):
		h.writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	default:
		h.writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		h.logger.Error("internal server error", "error", err)
	}
	return true
}

func (h *SubscriptionHandler) writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		h.logger.Error("failed to encode json response", "error", err)
	}
}
