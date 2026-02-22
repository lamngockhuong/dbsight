package adapter

import (
	"context"
	"fmt"
	"time"

	"github.com/lamngockhuong/dbsight/internal/models"
)

// GetSlowQueries fetches top slow queries from performance_schema.
func (a *MySQLAdapter) GetSlowQueries(ctx context.Context, opts QueryOpts) ([]models.SlowQuery, error) {
	limit := 50
	if opts.Limit > 0 {
		limit = opts.Limit
	}

	rows, err := a.db.QueryContext(ctx, `
		SELECT
			IFNULL(DIGEST, '') AS query_id,
			IFNULL(DIGEST_TEXT, '') AS query_text,
			COUNT_STAR AS calls,
			SUM_TIMER_WAIT / 1000000000 AS total_exec_ms,
			(SUM_TIMER_WAIT / COUNT_STAR) / 1000000000 AS mean_exec_ms,
			SUM_ROWS_EXAMINED AS rows_examined
		FROM performance_schema.events_statements_summary_by_digest
		WHERE SCHEMA_NAME NOT IN ('mysql', 'performance_schema', 'information_schema', 'sys')
		  AND COUNT_STAR > 0
		  AND (SUM_TIMER_WAIT / COUNT_STAR) / 1000000000 >= ?
		ORDER BY SUM_TIMER_WAIT DESC
		LIMIT ?
	`, opts.MinMeanMs, limit)
	if err != nil {
		return nil, fmt.Errorf("mysql slow queries: %w", err)
	}
	defer rows.Close()

	now := time.Now()
	var result []models.SlowQuery
	for rows.Next() {
		var q models.SlowQuery
		if err := rows.Scan(&q.QueryID, &q.Query, &q.Calls, &q.TotalExecMs, &q.MeanExecMs, &q.Rows); err != nil {
			return nil, fmt.Errorf("scan mysql slow query: %w", err)
		}
		q.SnapshotAt = now
		result = append(result, q)
	}
	return result, rows.Err()
}
