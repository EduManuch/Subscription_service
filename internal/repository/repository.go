package repository

import (
	"context"
	"fmt"
	"strings"
	"sub_service/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SubscriptionRepository struct {
	db *pgxpool.Pool
}

type ListSubscriptionsFilter struct {
	UserID      *uuid.UUID
	ServiceName *string
	Limit       int
	Offset      int
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

func (sr *SubscriptionRepository) List(ctx context.Context, f ListSubscriptionsFilter) ([]model.Subscription, error) {
	query := `
	SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
	FROM subscriptions
	`

	var args []any
	var conditions []string
	argPosition := 1

	if f.UserID != nil {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argPosition))
		args = append(args, *f.UserID)
		argPosition++
	}

	if f.ServiceName != nil && *f.ServiceName != "" {
		conditions = append(conditions, fmt.Sprintf("service_name = $%d", argPosition))
		args = append(args, *f.ServiceName)
		argPosition++
	}

	if len(conditions) > 0 {
		query += " WHERE " + fmt.Sprintf(strings.Join(conditions, " AND "))
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argPosition, argPosition+1)
	args = append(args, f.Limit, f.Offset)

	rows, err := sr.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	subscriptions := make([]model.Subscription, 0)

	for rows.Next() {
		var s model.Subscription
		err := rows.Scan(
			&s.ID,
			&s.ServiceName,
			&s.Price,
			&s.UserID,
			&s.StartDate,
			&s.EndDate,
			&s.CreatedAt,
			&s.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}
		subscriptions = append(subscriptions, s)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return subscriptions, nil
}
