package service

import (
	"context"
	"errors"
	"fmt"
	"sub_service/internal/model"
	"sub_service/internal/repository"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var (
	ErrSubNotFound  = errors.New("subscription not found")
	ErrInvalidSubID = errors.New("invalid subscription id")
)

type SubscriptionService struct {
	repo *repository.SubscriptionRepository
}

func NewSubscriptionService(repo *repository.SubscriptionRepository) *SubscriptionService {
	return &SubscriptionService{repo: repo}
}

type CreationSubscriptionInput struct {
	ServiceName string  `json:"service_name"`
	Price       int     `json:"price"`
	UserID      string  `json:"user_id"`
	StartDate   string  `json:"start_date"`
	EndDate     *string `json:"end_date,omitempty"`
}

func (s *SubscriptionService) Create(ctx context.Context, input CreationSubscriptionInput) (*model.Subscription, error) {
	if input.ServiceName == "" {
		return nil, fmt.Errorf("service name is required")
	}

	if input.Price <= 0 {
		return nil, fmt.Errorf("price must be greater than 0")
	}

	userID, err := uuid.Parse(input.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	startDate, err := time.Parse("01-2006", input.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start date: %w", err)
	}

	var endDate *time.Time
	if input.EndDate != nil && *input.EndDate != "" {
		parsedEndDate, err := time.Parse("01-2006", *input.EndDate)
		if err != nil {
			return nil, fmt.Errorf("invalid end date: %w", err)
		}
		endDate = &parsedEndDate
		if endDate.Before(startDate) {
			return nil, fmt.Errorf("invalid end date: should be equal or after start date")
		}
	}

	mSubscription := &model.Subscription{
		ServiceName: input.ServiceName,
		Price:       input.Price,
		UserID:      userID,
		StartDate:   startDate,
		EndDate:     endDate,
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
	for input.UserID != nil && *input.UserID == "" {
		parsed, err := uuid.Parse(*input.UserID)
		if err != nil {
			return nil, err
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
		return nil, fmt.Errorf("invalid subscription id: %w", err)
	}

	if input.ServiceName == "" {
		return nil, fmt.Errorf("service name is required")
	}

	if input.Price <= 0 {
		return nil, fmt.Errorf("price must be greater than 0")
	}

	userID, err := uuid.Parse(input.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	startDate, err := time.Parse("01-2006", input.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start date: %w", err)
	}

	var endDate *time.Time
	if input.EndDate != nil && *input.EndDate != "" {
		parsedEndDate, err := time.Parse("01-2006", *input.EndDate)
		if err != nil {
			return nil, fmt.Errorf("invalid end date: %w", err)
		}
		endDate = &parsedEndDate
		if endDate.Before(startDate) {
			return nil, fmt.Errorf("invalid end date: should be equal or after start date")
		}
	}

	mSubscription := &model.Subscription{
		ID:          parsedID,
		ServiceName: input.ServiceName,
		Price:       input.Price,
		UserID:      userID,
		StartDate:   startDate,
		EndDate:     endDate,
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
		return fmt.Errorf("invalid subscription id: %w", err)
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
