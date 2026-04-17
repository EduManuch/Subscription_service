package service

import (
	"context"
	"errors"
	"sub_service/internal/model"
	"sub_service/internal/repository"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var (
	ErrSubNotFound             = errors.New("subscription not found")
	ErrInvalidSubID            = errors.New("invalid subscription id")
	ErrServiceNameRequired     = errors.New("service name is required")
	ErrPriceLessThan0          = errors.New("price must be greater than 0")
	ErrInvalidUserID           = errors.New("invalid user ID")
	ErrInvalidStartDate        = errors.New("invalid start date")
	ErrInvalidEndDate          = errors.New("invalid end date")
	ErrEndDateGreaterStartDate = errors.New("end date must be equal or after start date")
)

type SubscriptionService struct {
	repo subscriptionRepository
}

// Интерфейс для использования в тестах
type subscriptionRepository interface {
	Create(ctx context.Context, s *model.Subscription) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error)
	List(ctx context.Context, f repository.ListSubscriptionsFilter) ([]model.Subscription, error)
	Update(ctx context.Context, s *model.Subscription) error
	Delete(ctx context.Context, id uuid.UUID) error
	CalculateTotalPrice(ctx context.Context, f repository.CalculatePriceFilter) (int, error)
}

func NewSubscriptionService(repo subscriptionRepository) *SubscriptionService {
	return &SubscriptionService{repo: repo}
}

type ValidationInput struct {
	ServiceName string
	Price       int
	UserID      string
	StartDate   string
	EndDate     *string
}

type ValidatedData struct {
	UserID    uuid.UUID
	StartDate time.Time
	EndDate   *time.Time
}

type CreationSubscriptionInput struct {
	ServiceName string  `json:"service_name"`
	Price       int     `json:"price"`
	UserID      string  `json:"user_id"`
	StartDate   string  `json:"start_date"`
	EndDate     *string `json:"end_date,omitempty"`
}

func (s *SubscriptionService) Create(ctx context.Context, input CreationSubscriptionInput) (*model.Subscription, error) {
	vInput, err := validateInput(&ValidationInput{
		ServiceName: input.ServiceName,
		Price:       input.Price,
		UserID:      input.UserID,
		StartDate:   input.StartDate,
		EndDate:     input.EndDate,
	})

	if err != nil {
		return nil, err
	}

	mSubscription := &model.Subscription{
		ServiceName: input.ServiceName,
		Price:       input.Price,
		UserID:      vInput.UserID,
		StartDate:   vInput.StartDate,
		EndDate:     vInput.EndDate,
	}

	if err := s.repo.Create(ctx, mSubscription); err != nil {
		return nil, err
	}

	return mSubscription, nil
}

func (s *SubscriptionService) GetByID(ctx context.Context, id string) (*model.Subscription, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrInvalidSubID
	}

	sub, err := s.repo.GetByID(ctx, parsedID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSubNotFound
		}
		return nil, err
	}
	return sub, nil
}

type ListSubscriptionsInput struct {
	UserID      *string
	ServiceName *string
	Limit       int
	Offset      int
}

func (s *SubscriptionService) List(ctx context.Context, input ListSubscriptionsInput) ([]model.Subscription, error) {
	limit := input.Limit
	offset := input.Offset

	// Учтем, что таблица может быть очень большой и получать все записи разом будет неоптимально
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	var userID *uuid.UUID
	if input.UserID != nil && *input.UserID != "" {
		parsed, err := uuid.Parse(*input.UserID)
		if err != nil {
			return nil, ErrInvalidSubID
		}
		userID = &parsed
	}

	filter := repository.ListSubscriptionsFilter{
		UserID:      userID,
		ServiceName: input.ServiceName,
		Limit:       limit,
		Offset:      offset,
	}

	return s.repo.List(ctx, filter)
}

type UpdateSubscriptionInput struct {
	ServiceName string  `json:"service_name"`
	Price       int     `json:"price"`
	UserID      string  `json:"user_id"`
	StartDate   string  `json:"start_date"`
	EndDate     *string `json:"end_date,omitempty"`
}

func (s *SubscriptionService) Update(ctx context.Context, id string, input UpdateSubscriptionInput) (*model.Subscription, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrInvalidSubID
	}

	vInput, err := validateInput(&ValidationInput{
		ServiceName: input.ServiceName,
		Price:       input.Price,
		UserID:      input.UserID,
		StartDate:   input.StartDate,
		EndDate:     input.EndDate,
	})

	if err != nil {
		return nil, err
	}

	mSubscription := &model.Subscription{
		ID:          parsedID,
		ServiceName: input.ServiceName,
		Price:       input.Price,
		UserID:      vInput.UserID,
		StartDate:   vInput.StartDate,
		EndDate:     vInput.EndDate,
	}

	err = s.repo.Update(ctx, mSubscription)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSubNotFound
		}
		return nil, err
	}

	return mSubscription, nil
}

func (s *SubscriptionService) Delete(ctx context.Context, id string) error {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return ErrInvalidSubID
	}

	err = s.repo.Delete(ctx, parsedID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrSubNotFound
		}
		return err
	}

	return nil
}

type CalculateTotalPriceInput struct {
	From        string
	To          string
	UserID      *string
	ServiceName *string
}

func (s *SubscriptionService) CalculateTotalPrice(ctx context.Context, input CalculateTotalPriceInput) (int, error) {
	from, err := time.Parse("01-2006", input.From)
	if err != nil {
		return 0, ErrInvalidStartDate
	}

	to, err := time.Parse("01-2006", input.To)
	if err != nil {
		return 0, ErrInvalidEndDate
	}

	if to.Before(from) {
		return 0, ErrEndDateGreaterStartDate
	}

	var userID *uuid.UUID
	if input.UserID != nil && *input.UserID != "" {
		parsed, err := uuid.Parse(*input.UserID)
		if err != nil {
			return 0, ErrInvalidUserID
		}
		userID = &parsed
	}

	filter := repository.CalculatePriceFilter{
		StartDate:   from,
		EndDate:     to,
		UserID:      userID,
		ServiceName: input.ServiceName,
	}

	return s.repo.CalculateTotalPrice(ctx, filter)
}

func validateInput(input *ValidationInput) (*ValidatedData, error) {

	if input.ServiceName == "" {
		return nil, ErrServiceNameRequired
	}

	if input.Price <= 0 {
		return nil, ErrPriceLessThan0
	}

	userID, err := uuid.Parse(input.UserID)
	if err != nil {
		return nil, ErrInvalidUserID
	}

	startDate, err := time.Parse("01-2006", input.StartDate)
	if err != nil {
		return nil, ErrInvalidStartDate
	}

	var endDate *time.Time
	if input.EndDate != nil && *input.EndDate != "" {
		parsedEndDate, err := time.Parse("01-2006", *input.EndDate)
		if err != nil {
			return nil, ErrInvalidEndDate
		}
		endDate = &parsedEndDate
		if endDate.Before(startDate) {
			return nil, ErrEndDateGreaterStartDate
		}
	}

	result := &ValidatedData{
		UserID:    userID,
		StartDate: startDate,
		EndDate:   endDate,
	}
	return result, nil
}
