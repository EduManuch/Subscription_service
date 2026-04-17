package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"sub_service/internal/model"
	"sub_service/internal/repository"
	serv "sub_service/internal/service"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
)

type mockSubscriptionRepository struct {
	createFn              func(ctx context.Context, s *model.Subscription) error
	getByIDFn             func(ctx context.Context, id uuid.UUID) (*model.Subscription, error)
	listFn                func(ctx context.Context, f repository.ListSubscriptionsFilter) ([]model.Subscription, error)
	updateFn              func(ctx context.Context, s *model.Subscription) error
	deleteFn              func(ctx context.Context, id uuid.UUID) error
	calculateTotalPriceFn func(ctx context.Context, f repository.CalculatePriceFilter) (int, error)
}

func (m *mockSubscriptionRepository) Create(ctx context.Context, s *model.Subscription) error {
	if m.createFn == nil {
		return nil
	}
	return m.createFn(ctx, s)
}

func (m *mockSubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	if m.getByIDFn == nil {
		return nil, nil
	}
	return m.getByIDFn(ctx, id)
}

func (m *mockSubscriptionRepository) List(ctx context.Context, f repository.ListSubscriptionsFilter) ([]model.Subscription, error) {
	if m.listFn == nil {
		return nil, nil
	}
	return m.listFn(ctx, f)
}

func (m *mockSubscriptionRepository) Update(ctx context.Context, s *model.Subscription) error {
	if m.updateFn == nil {
		return nil
	}
	return m.updateFn(ctx, s)
}

func (m *mockSubscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFn == nil {
		return nil
	}
	return m.deleteFn(ctx, id)
}

func (m *mockSubscriptionRepository) CalculateTotalPrice(ctx context.Context, f repository.CalculatePriceFilter) (int, error) {
	if m.calculateTotalPriceFn == nil {
		return 0, nil
	}
	return m.calculateTotalPriceFn(ctx, f)
}

func TestCreate_InvalidBody(t *testing.T) {
	t.Parallel()

	h := newTestHandler(&mockSubscriptionRepository{})
	req := httptest.NewRequest(http.MethodPost, "/subscriptions", strings.NewReader("{invalid-json"))
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	var resp ErrorResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "invalid request body", resp.Error)
}

func TestCreate_Success(t *testing.T) {
	t.Parallel()

	date := time.Date(2026, time.January, 2, 10, 0, 0, 0, time.UTC)
	userID := uuid.New()

	h := newTestHandler(&mockSubscriptionRepository{
		createFn: func(_ context.Context, s *model.Subscription) error {
			s.ID = uuid.New()
			s.CreatedAt = date
			s.UpdatedAt = date
			return nil
		},
	})

	body := `{"service_name":"Netflix",
		"price":999,"user_id":"` +
		userID.String() + `",
		"start_date":"01-2025",
		"end_date":"02-2025"}`
	req := httptest.NewRequest(http.MethodPost, "/subscriptions", strings.NewReader(body))
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	var resp CRLResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "Netflix", resp.ServiceName)
	assert.Equal(t, 999, resp.Price)
	assert.Equal(t, userID, resp.UserID)
	assert.Equal(t, "01-2025", resp.StartDate)
	assert.Equal(t, "02-2025", resp.EndDate)
}

func TestGetByID_InvalidID(t *testing.T) {
	t.Parallel()

	h := newTestHandler(&mockSubscriptionRepository{})
	req := httptest.NewRequest(http.MethodGet, "/subscriptions/incorrect-id", nil)
	req.SetPathValue("id", "incorrect-id")
	rec := httptest.NewRecorder()

	h.GetByID(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetByID_NotFound(t *testing.T) {
	t.Parallel()

	h := newTestHandler(&mockSubscriptionRepository{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*model.Subscription, error) {
			return nil, pgx.ErrNoRows
		},
	})
	req := httptest.NewRequest(http.MethodGet, "/subscriptions/1", nil)
	req.SetPathValue("id", uuid.New().String())
	rec := httptest.NewRecorder()

	h.GetByID(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestList_InvalidLimit(t *testing.T) {
	t.Parallel()

	h := newTestHandler(&mockSubscriptionRepository{})
	req := httptest.NewRequest(http.MethodGet, "/subscriptions?limit=abc", nil)
	rec := httptest.NewRecorder()

	h.List(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	var resp ErrorResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "invalid limit", resp.Error)
}

func TestList_Success(t *testing.T) {
	t.Parallel()

	date := time.Date(2026, time.April, 1, 10, 0, 0, 0, time.UTC)
	userID := uuid.New()
	subID := uuid.New()

	h := newTestHandler(&mockSubscriptionRepository{
		listFn: func(_ context.Context, f repository.ListSubscriptionsFilter) ([]model.Subscription, error) {
			assert.Equal(t, 15, f.Limit)
			assert.Equal(t, 2, f.Offset)
			assert.NotNil(t, f.ServiceName)
			assert.Equal(t, "Netflix", *f.ServiceName)

			return []model.Subscription{
				{
					ID:          subID,
					ServiceName: "Netflix",
					Price:       500,
					UserID:      userID,
					StartDate:   time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
					CreatedAt:   date,
					UpdatedAt:   date,
				},
			}, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/subscriptions?service_name=Netflix&limit=15&offset=2", nil)
	rec := httptest.NewRecorder()

	h.List(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp []CRLResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Len(t, resp, 1)
	assert.Equal(t, subID, resp[0].ID)
	assert.Equal(t, "Netflix", resp[0].ServiceName)
	assert.Equal(t, "01-2026", resp[0].StartDate)
}

func TestUpdate_InvalidBody(t *testing.T) {
	t.Parallel()

	h := newTestHandler(&mockSubscriptionRepository{})
	req := httptest.NewRequest(http.MethodPut, "/subscriptions/1", strings.NewReader("{incorrect"))
	req.SetPathValue("id", uuid.New().String())
	rec := httptest.NewRecorder()

	h.Update(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestDelete_Success(t *testing.T) {
	t.Parallel()

	h := newTestHandler(&mockSubscriptionRepository{})
	req := httptest.NewRequest(http.MethodDelete, "/subscriptions/1", nil)
	req.SetPathValue("id", uuid.New().String())
	rec := httptest.NewRecorder()

	h.Delete(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestCalculateTotalPrice_EmptyDates(t *testing.T) {
	t.Parallel()

	h := newTestHandler(&mockSubscriptionRepository{})
	req := httptest.NewRequest(http.MethodGet, "/subscriptions/total?from=&to=", nil)
	rec := httptest.NewRecorder()

	h.CalculateTotalPrice(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	var resp ErrorResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "from and to dates are required", resp.Error)
}

func TestCalculateTotalPrice_Success(t *testing.T) {
	t.Parallel()

	h := newTestHandler(&mockSubscriptionRepository{
		calculateTotalPriceFn: func(_ context.Context, f repository.CalculatePriceFilter) (int, error) {
			assert.Equal(t, time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC), f.StartDate)
			assert.Equal(t, time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC), f.EndDate)
			return 1234, nil
		},
	})
	req := httptest.NewRequest(http.MethodGet, "/subscriptions/total?from=01-2026&to=03-2026", nil)
	rec := httptest.NewRecorder()

	h.CalculateTotalPrice(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp TotalPriceResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 1234, resp.Total)
	assert.Equal(t, "01-2026", resp.From)
	assert.Equal(t, "03-2026", resp.To)
}

func TestResponseError_InternalError(t *testing.T) {
	t.Parallel()

	h := newTestHandler(&mockSubscriptionRepository{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*model.Subscription, error) {
			return nil, errors.New("db is down")
		},
	})
	req := httptest.NewRequest(http.MethodGet, "/subscriptions/1", nil)
	req.SetPathValue("id", uuid.New().String())
	rec := httptest.NewRecorder()

	h.GetByID(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	var resp ErrorResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "internal server error", resp.Error)
}

func newTestHandler(repo *mockSubscriptionRepository) *SubscriptionHandler {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	svc := serv.NewSubscriptionService(repo)
	return NewSubscriptionHandler(svc, logger)
}
