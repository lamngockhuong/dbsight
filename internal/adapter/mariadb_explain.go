package adapter

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/lamngockhuong/dbsight/internal/models"
)

// GetExplainPlan returns the query execution plan for MariaDB.
// EXPLAIN FORMAT=JSON for estimates; ANALYZE FORMAT=JSON for SELECT (10.1+).
func (a *MariaDBAdapter) GetExplainPlan(ctx context.Context, query string, opts QueryOpts) (*models.ExplainPlan, error) {
	if opts.AnalyzeMode && !isSelectQuery(query) {
		return nil, fmt.Errorf("ANALYZE only supported for SELECT queries")
	}

	tx, err := a.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	explainSQL := "EXPLAIN FORMAT=JSON " + query
	if opts.AnalyzeMode {
		// MariaDB uses ANALYZE FORMAT=JSON (returns JSON, not TREE like MySQL)
		explainSQL = "ANALYZE FORMAT=JSON " + query
	}

	var planRaw string
	if err := tx.QueryRowContext(ctx, explainSQL).Scan(&planRaw); err != nil {
		return nil, fmt.Errorf("explain: %w", err)
	}
	return &models.ExplainPlan{
		QueryText: query,
		PlanJSON:  json.RawMessage(planRaw),
	}, nil
}
