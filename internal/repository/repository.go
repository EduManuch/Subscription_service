package repository

import (
	"context"
	"sub_service/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SubscriptionRepository struct {
	db *pgxpool.Pool
}

func NewSubscriptionRepository(db *pgxpool.Pool) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

func (sr *SubscriptionRepository) Create(ctx context.Context, s *model.Subscription) error {
	query := `
	INSERT INTO subscriptions (
		service_name,
	    price,
	    user_id,
	    start_date,
	    end_date
	)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id, created_at, updated_at
	`

	return sr.db.QueryRow(
		ctx,
		query,
		s.ServiceName,
		s.Price,
		s.UserID,
		s.StartDate,
		s.EndDate,
	).Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)
}

func (sr *SubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	query := `
	SELECT 
		id,
		service_name,
		price,
		user_id,
		start_date,
		end_date,
		created_at,
		updated_at
	FROM subscriptions
	WHERE id = $1
	`

	var sub model.Subscription
	err := sr.db.QueryRow(ctx, query, id).Scan(
		&sub.ID,
		&sub.ServiceName,
		&sub.Price,
		&sub.UserID,
		&sub.StartDate,
		&sub.EndDate,
		&sub.CreatedAt,
		&sub.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}
	return &sub, nil
}
