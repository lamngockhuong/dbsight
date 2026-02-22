package adapter

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/lamngockhuong/dbsight/internal/models"
)

// GetExplainPlan returns the query execution plan.
// Uses a read-only transaction to prevent SQL injection side effects.
func (a *PostgresAdapter) GetExplainPlan(ctx context.Context, query string, opts QueryOpts) (*models.ExplainPlan, error) {
	explainSQL := "EXPLAIN (FORMAT JSON) " + query
	if opts.AnalyzeMode {
		explainSQL = "EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON) " + query
	}

	tx, err := a.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Force read-only to mitigate SQL injection risk
	if _, err := tx.Exec(ctx, "SET TRANSACTION READ ONLY"); err != nil {
		return nil, fmt.Errorf("set read only: %w", err)
	}

	var planJSON json.RawMessage
	if err := tx.QueryRow(ctx, explainSQL).Scan(&planJSON); err != nil {
		return nil, fmt.Errorf("explain: %w", err)
	}
	return &models.ExplainPlan{QueryText: query, PlanJSON: planJSON}, nil
}
