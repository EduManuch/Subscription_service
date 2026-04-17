package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"sub_service/internal/model"
	"sub_service/internal/repository"

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

func TestCreate_ValidInput(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	repo := &mockSubscriptionRepository{
		createFn: func(_ context.Context, s *model.Subscription) error {
			assert.Equal(t, "Netflix", s.ServiceName)
			assert.Equal(t, 999, s.Price)
			assert.Equal(t, userID, s.UserID)
			assert.Equal(t, time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC), s.StartDate)
			assert.Nil(t, s.EndDate)

			return nil
		},
	}

	svc := NewSubscriptionService(repo)
	sub, err := svc.Create(context.Background(), CreationSubscriptionInput{
		ServiceName: "Netflix",
		Price:       999,
		UserID:      userID.String(),
		StartDate:   "01-2025",
	})

	assert.NoError(t, err)
	assert.NotNil(t, sub)
}

func TestCreate_InvalidInput(t *testing.T) {
	t.Parallel()

	svc := NewSubscriptionService(&mockSubscriptionRepository{})
	_, err := svc.Create(context.Background(), CreationSubscriptionInput{
		ServiceName: "",
		Price:       10,
		UserID:      uuid.New().String(),
		StartDate:   "01-2025",
	})

	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrServiceNameRequired))
}

func TestGetByID_InvalidID(t *testing.T) {
	t.Parallel()

	svc := NewSubscriptionService(&mockSubscriptionRepository{})
	_, err := svc.GetByID(context.Background(), "incorrect-id")

	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidSubID))
}

func TestGetByID_NotFound(t *testing.T) {
	t.Parallel()

	svc := NewSubscriptionService(&mockSubscriptionRepository{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*model.Subscription, error) {
			return nil, pgx.ErrNoRows
		},
	})

	_, err := svc.GetByID(context.Background(), uuid.New().String())

	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrSubNotFound))
}

func TestList_LimitOffsetControl(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	serviceName := "Romashka"
	var controlled repository.ListSubscriptionsFilter

	svc := NewSubscriptionService(&mockSubscriptionRepository{
		listFn: func(_ context.Context, f repository.ListSubscriptionsFilter) ([]model.Subscription, error) {
			controlled = f
			return []model.Subscription{}, nil
		},
	})

	_, err := svc.List(context.Background(), ListSubscriptionsInput{
		UserID:      ptr(userID.String()),
		ServiceName: &serviceName,
		Limit:       1000,
		Offset:      -5,
	})

	assert.NoError(t, err)
	assert.Equal(t, 100, controlled.Limit)
	assert.Equal(t, 0, controlled.Offset)
}

func TestList_InvalidUserID(t *testing.T) {
	t.Parallel()

	svc := NewSubscriptionService(&mockSubscriptionRepository{})
	_, err := svc.List(context.Background(), ListSubscriptionsInput{
		UserID: ptr("incorrect-id"),
		Limit:  10,
	})

	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidSubID))
}

func TestUpdate_NotFound(t *testing.T) {
	t.Parallel()

	svc := NewSubscriptionService(&mockSubscriptionRepository{
		updateFn: func(_ context.Context, _ *model.Subscription) error {
			return pgx.ErrNoRows
		},
	})

	_, err := svc.Update(context.Background(), uuid.New().String(), UpdateSubscriptionInput{
		ServiceName: "Netflix",
		Price:       100,
		UserID:      uuid.New().String(),
		StartDate:   "01-2025",
	})

	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrSubNotFound))
}

func TestDelete_InvalidID(t *testing.T) {
	t.Parallel()

	svc := NewSubscriptionService(&mockSubscriptionRepository{})
	err := svc.Delete(context.Background(), "incorrect-id")

	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidSubID))
}

func TestDelete_NotFound(t *testing.T) {
	t.Parallel()

	svc := NewSubscriptionService(&mockSubscriptionRepository{
		deleteFn: func(_ context.Context, _ uuid.UUID) error {
			return pgx.ErrNoRows
		},
	})

	err := svc.Delete(context.Background(), uuid.New().String())

	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrSubNotFound))
}

func TestCalculateTotalPrice_InvalidFromDate(t *testing.T) {
	t.Parallel()

	svc := NewSubscriptionService(&mockSubscriptionRepository{})
	_, err := svc.CalculateTotalPrice(context.Background(), CalculateTotalPriceInput{
		From: "incorrect",
		To:   "02-2025",
	})

	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidStartDate))
}

func TestCalculateTotalPrice_FromAfterTo(t *testing.T) {
	t.Parallel()

	svc := NewSubscriptionService(&mockSubscriptionRepository{})
	_, err := svc.CalculateTotalPrice(context.Background(), CalculateTotalPriceInput{
		From: "03-2025",
		To:   "02-2025",
	})

	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrEndDateGreaterStartDate))
}

func TestCalculateTotalPrice_ValidInput(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	serviceName := "Romashka"
	var controlled repository.CalculatePriceFilter

	svc := NewSubscriptionService(&mockSubscriptionRepository{
		calculateTotalPriceFn: func(_ context.Context, f repository.CalculatePriceFilter) (int, error) {
			controlled = f
			return 1200, nil
		},
	})

	total, err := svc.CalculateTotalPrice(context.Background(), CalculateTotalPriceInput{
		From:        "01-2025",
		To:          "03-2025",
		UserID:      ptr(userID.String()),
		ServiceName: &serviceName,
	})

	assert.NoError(t, err)
	assert.Equal(t, 1200, total)
	assert.Equal(t, time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC), controlled.StartDate)
	assert.Equal(t, time.Date(2025, time.March, 1, 0, 0, 0, 0, time.UTC), controlled.EndDate)
	assert.NotNil(t, controlled.UserID)
	assert.Equal(t, userID, *controlled.UserID)
	assert.NotNil(t, controlled.ServiceName)
	assert.Equal(t, serviceName, *controlled.ServiceName)
}

// Получаем указатель на литерал
func ptr(v string) *string {
	return &v
}
