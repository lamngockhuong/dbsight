package adapter

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresAdapter implements DBAnalyzer for PostgreSQL.
type PostgresAdapter struct {
	pool *pgxpool.Pool
}

// NewPostgresAdapter creates an unconnected PostgresAdapter.
func NewPostgresAdapter() *PostgresAdapter {
	return &PostgresAdapter{}
}

// Connect opens a connection pool to the target database and verifies connectivity.
func (a *PostgresAdapter) Connect(ctx context.Context, dsn string) error {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return fmt.Errorf("connect target db: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return fmt.Errorf("ping target db: %w", err)
	}
	a.pool = pool
	return nil
}

// Close releases the connection pool.
func (a *PostgresAdapter) Close() error {
	if a.pool != nil {
		a.pool.Close()
	}
	return nil
}
