package service

import (
	"context"
	"fmt"
	"sub_service/internal/model"
	"sub_service/internal/repository"
	"time"

	"github.com/google/uuid"
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
