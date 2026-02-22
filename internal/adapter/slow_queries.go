package adapter

import (
	"context"
	"fmt"
	"time"

	"github.com/lamngockhuong/dbsight/internal/models"
)

// GetSlowQueries fetches top slow queries from pg_stat_statements.
func (a *PostgresAdapter) GetSlowQueries(ctx context.Context, opts QueryOpts) ([]models.SlowQuery, error) {
	limit := 50
	if opts.Limit > 0 {
		limit = opts.Limit
	}

	rows, err := a.pool.Query(ctx, `
		SELECT
			queryid::text,
			query,
			calls,
			total_exec_time,
			mean_exec_time,
			rows
		FROM pg_stat_statements
		WHERE mean_exec_time >= $1
		ORDER BY total_exec_time DESC
		LIMIT $2
	`, opts.MinMeanMs, limit)
	if err != nil {
		return nil, fmt.Errorf("pg_stat_statements query: %w (ensure pg_stat_statements extension is enabled)", err)
	}
	defer rows.Close()

	now := time.Now()
	var result []models.SlowQuery
	for rows.Next() {
		var q models.SlowQuery
		if err := rows.Scan(&q.QueryID, &q.Query, &q.Calls, &q.TotalExecMs, &q.MeanExecMs, &q.Rows); err != nil {
			return nil, fmt.Errorf("scan slow query: %w", err)
		}
		q.SnapshotAt = now
		result = append(result, q)
	}
	return result, rows.Err()
}
