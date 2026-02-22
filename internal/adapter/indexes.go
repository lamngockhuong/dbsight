package adapter

import (
	"context"
	"fmt"

	"github.com/lamngockhuong/dbsight/internal/models"
)

// GetIndexStats returns per-index usage stats from pg_stat_user_indexes.
func (a *PostgresAdapter) GetIndexStats(ctx context.Context) ([]models.IndexStat, error) {
	rows, err := a.pool.Query(ctx, `
		SELECT
			schemaname,
			relname,
			indexrelname,
			idx_scan,
			idx_tup_read,
			idx_tup_fetch,
			pg_relation_size(indexrelid) AS index_size_bytes
		FROM pg_stat_user_indexes
		ORDER BY idx_scan ASC, pg_relation_size(indexrelid) DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("index stats: %w", err)
	}
	defer rows.Close()

	var result []models.IndexStat
	for rows.Next() {
		var s models.IndexStat
		if err := rows.Scan(
			&s.SchemaName, &s.TableName, &s.IndexName,
			&s.IndexScans, &s.TupRead, &s.TupFetch, &s.IndexSizeB,
		); err != nil {
			return nil, fmt.Errorf("scan index stat: %w", err)
		}
		s.IsUnused = s.IndexScans == 0
		result = append(result, s)
	}
	return result, rows.Err()
}

// GetTableStats returns per-table usage stats from pg_stat_user_tables.
func (a *PostgresAdapter) GetTableStats(ctx context.Context) ([]models.TableStat, error) {
	rows, err := a.pool.Query(ctx, `
		SELECT
			schemaname,
			relname,
			seq_scan,
			seq_tup_read,
			n_live_tup,
			pg_total_relation_size(relid) AS table_size_bytes
		FROM pg_stat_user_tables
		ORDER BY seq_scan DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("table stats: %w", err)
	}
	defer rows.Close()

	var result []models.TableStat
	for rows.Next() {
		var s models.TableStat
		if err := rows.Scan(
			&s.SchemaName, &s.TableName, &s.SeqScans,
			&s.SeqTupRead, &s.NLiveTup, &s.TableSizeB,
		); err != nil {
			return nil, fmt.Errorf("scan table stat: %w", err)
		}
		result = append(result, s)
	}
	return result, rows.Err()
}
