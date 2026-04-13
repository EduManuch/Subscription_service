package storage

import (
	"context"
	"fmt"
	"log"
	"sub_service/internal/config"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgres(cfg *config.Config) (*pgxpool.Pool, error) {
	dataSource := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dbPool, err := pgxpool.New(ctx, dataSource)
	if err != nil {
		return nil, err
	}

	if err = dbPool.Ping(ctx); err != nil {
		return nil, err
	}

	log.Println("Connected to PotgreSQL")
	return dbPool, nil
}
