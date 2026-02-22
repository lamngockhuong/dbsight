package adapter

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/lamngockhuong/dbsight/internal/models"
)

// GetExplainPlan returns the query execution plan for MySQL.
// EXPLAIN FORMAT=JSON for estimates; EXPLAIN ANALYZE for SELECT on 8.0.18+.
func (a *MySQLAdapter) GetExplainPlan(ctx context.Context, query string, opts QueryOpts) (*models.ExplainPlan, error) {
	if opts.AnalyzeMode {
		if !isSelectQuery(query) {
			return nil, fmt.Errorf("ANALYZE only supported for SELECT queries")
		}
		if !a.version.AtLeastPatch(8, 0, 18) {
			return nil, fmt.Errorf("EXPLAIN ANALYZE requires MySQL 8.0.18+")
		}
	}

	tx, err := a.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	explainSQL := "EXPLAIN FORMAT=JSON " + query
	if opts.AnalyzeMode {
		explainSQL = "EXPLAIN ANALYZE " + query
	}

	var planRaw string
	if err := tx.QueryRowContext(ctx, explainSQL).Scan(&planRaw); err != nil {
		return nil, fmt.Errorf("explain: %w", err)
	}

	// ANALYZE returns TREE text (not JSON); wrap in JSON envelope
	if opts.AnalyzeMode {
		return &models.ExplainPlan{
			QueryText: query,
			PlanJSON:  json.RawMessage(fmt.Sprintf(`{"format":"tree","text":%s}`, strconv.Quote(planRaw))),
		}, nil
	}
	return &models.ExplainPlan{
		QueryText: query,
		PlanJSON:  json.RawMessage(planRaw),
	}, nil
}
