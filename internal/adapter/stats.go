package adapter

import (
	"context"
	"fmt"

	"github.com/lamngockhuong/dbsight/internal/models"
)

// GetDatabaseStats returns high-level database metrics.
func (a *PostgresAdapter) GetDatabaseStats(ctx context.Context) (*models.DatabaseStats, error) {
	var s models.DatabaseStats
	var cacheHitRatio *float64
	err := a.pool.QueryRow(ctx, `
		SELECT
			current_database(),
			pg_database_size(current_database()),
			(SELECT count(*) FROM pg_stat_activity WHERE state = 'active'),
			current_setting('max_connections')::int,
			ROUND(
				sum(blks_hit)::numeric /
				NULLIF(sum(blks_hit) + sum(blks_read), 0) * 100, 2
			)
		FROM pg_stat_database WHERE datname = current_database()
	`).Scan(&s.DBName, &s.SizeBytes, &s.ActiveConns, &s.MaxConns, &cacheHitRatio)
	if err != nil {
		return nil, fmt.Errorf("database stats: %w", err)
	}
	if cacheHitRatio != nil {
		s.CacheHitRatio = *cacheHitRatio
	}
	return &s, nil
}
