package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// HealthRepository provides the persistence check used by readiness probes.
type HealthRepository interface {
	Ping(context.Context) error
}

// PostgresHealthRepository is the PostgreSQL adapter for HealthRepository.
type PostgresHealthRepository struct {
	db *pgxpool.Pool
}

func NewHealthRepository(db *pgxpool.Pool) *PostgresHealthRepository {
	return &PostgresHealthRepository{db: db}
}

func (r *PostgresHealthRepository) Ping(ctx context.Context) error {
	return r.db.Ping(ctx)
}
