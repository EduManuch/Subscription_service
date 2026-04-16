package handlers

import (
	"time"

	"github.com/google/uuid"
)

// CRLResponse - Create, Read, List response
// Ответ для Create, GetByID и List
type CRLResponse struct {
	ID          uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	ServiceName string    `json:"service_name" example:"Netflix"`
	Price       int       `json:"price" example:"999"`
	UserID      uuid.UUID `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	StartDate   string    `json:"start_date" example:"01-2025"`
	EndDate     string    `json:"end_date,omitempty" example:"12-2025"`
	CreatedAt   time.Time `json:"created_at" example:"2025-01-01T00:00:00Z"`
	UpdatedAt   time.Time `json:"updated_at" example:"2025-01-02T00:00:00Z"`
}

type TotalPriceResponse struct {
	Total int    `json:"total" example:"1200"`
	From  string `json:"from" example:"01-2025"`
	To    string `json:"to" example:"12-2025"`
}

type ErrorResponse struct {
	Error string `json:"error" example:"invalid request body"`
}
