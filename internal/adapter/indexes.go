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

// GetDuplicateIndexes finds indexes with identical definitions on the same table.
func (a *PostgresAdapter) GetDuplicateIndexes(ctx context.Context) ([]models.DuplicateIndex, error) {
	rows, err := a.pool.Query(ctx, `
		SELECT
			a.tablename,
			a.indexname AS index1,
			b.indexname AS index2,
			a.indexdef
		FROM pg_indexes a
		JOIN pg_indexes b
			ON a.tablename = b.tablename
			AND a.indexdef = b.indexdef
			AND a.indexname < b.indexname
		WHERE a.schemaname NOT IN ('pg_catalog', 'information_schema')
		ORDER BY a.tablename, a.indexname
	`)
	if err != nil {
		return nil, fmt.Errorf("duplicate indexes: %w", err)
	}
	defer rows.Close()

	var result []models.DuplicateIndex
	for rows.Next() {
		var d models.DuplicateIndex
		if err := rows.Scan(&d.TableName, &d.Index1, &d.Index2, &d.IndexDef); err != nil {
			return nil, fmt.Errorf("scan duplicate index: %w", err)
		}
		result = append(result, d)
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
