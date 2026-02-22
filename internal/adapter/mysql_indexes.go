package adapter

import (
	"context"
	"fmt"
	"strings"

	"github.com/lamngockhuong/dbsight/internal/models"
)

// GetIndexStats returns unused indexes from MySQL.
// Tries sys.schema_unused_indexes first, falls back to performance_schema.
func (a *MySQLAdapter) GetIndexStats(ctx context.Context) ([]models.IndexStat, error) {
	result, err := a.getUnusedFromSys(ctx)
	if err != nil {
		result, err = a.getUnusedFromPerfSchema(ctx)
		if err != nil {
			return nil, fmt.Errorf("mysql index stats: %w", err)
		}
	}
	return result, nil
}

func (a *MySQLAdapter) getUnusedFromSys(ctx context.Context) ([]models.IndexStat, error) {
	rows, err := a.db.QueryContext(ctx, `
		SELECT object_schema, object_name, index_name
		FROM sys.schema_unused_indexes
		WHERE object_schema NOT IN ('mysql', 'performance_schema', 'information_schema', 'sys')
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.IndexStat
	for rows.Next() {
		var s models.IndexStat
		if err := rows.Scan(&s.SchemaName, &s.TableName, &s.IndexName); err != nil {
			return nil, fmt.Errorf("scan unused index: %w", err)
		}
		s.IsUnused = true
		result = append(result, s)
	}
	return result, rows.Err()
}

func (a *MySQLAdapter) getUnusedFromPerfSchema(ctx context.Context) ([]models.IndexStat, error) {
	rows, err := a.db.QueryContext(ctx, `
		SELECT OBJECT_SCHEMA, OBJECT_NAME, INDEX_NAME
		FROM performance_schema.table_io_waits_summary_by_index_usage
		WHERE INDEX_NAME IS NOT NULL AND COUNT_STAR = 0
		  AND OBJECT_SCHEMA NOT IN ('mysql', 'performance_schema', 'information_schema', 'sys')
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.IndexStat
	for rows.Next() {
		var s models.IndexStat
		if err := rows.Scan(&s.SchemaName, &s.TableName, &s.IndexName); err != nil {
			return nil, fmt.Errorf("scan unused index: %w", err)
		}
		s.IsUnused = true
		result = append(result, s)
	}
	return result, rows.Err()
}

// GetDuplicateIndexes finds indexes with overlapping column prefixes.
func (a *MySQLAdapter) GetDuplicateIndexes(ctx context.Context) ([]models.DuplicateIndex, error) {
	rows, err := a.db.QueryContext(ctx, `
		SELECT TABLE_SCHEMA, TABLE_NAME, INDEX_NAME,
		       GROUP_CONCAT(COLUMN_NAME ORDER BY SEQ_IN_INDEX) AS cols
		FROM INFORMATION_SCHEMA.STATISTICS
		WHERE TABLE_SCHEMA NOT IN ('mysql', 'performance_schema', 'information_schema', 'sys')
		  AND NON_UNIQUE = 1
		GROUP BY TABLE_SCHEMA, TABLE_NAME, INDEX_NAME
		ORDER BY TABLE_SCHEMA, TABLE_NAME
	`)
	if err != nil {
		return nil, fmt.Errorf("mysql duplicate indexes: %w", err)
	}
	defer rows.Close()

	type idxEntry struct {
		schema, table, name, cols string
	}
	var entries []idxEntry
	for rows.Next() {
		var e idxEntry
		if err := rows.Scan(&e.schema, &e.table, &e.name, &e.cols); err != nil {
			return nil, fmt.Errorf("scan index entry: %w", err)
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Find prefix duplicates within same table
	var result []models.DuplicateIndex
	for i, a := range entries {
		for j := i + 1; j < len(entries); j++ {
			b := entries[j]
			if a.schema != b.schema || a.table != b.table {
				continue
			}
			if strings.HasPrefix(a.cols, b.cols) || strings.HasPrefix(b.cols, a.cols) {
				result = append(result, models.DuplicateIndex{
					TableName: a.table,
					Index1:    a.name,
					Index2:    b.name,
					IndexDef:  fmt.Sprintf("%s vs %s", a.cols, b.cols),
				})
			}
		}
	}
	return result, nil
}

// GetTableStats returns tables with high sequential (full) scan counts.
func (a *MySQLAdapter) GetTableStats(ctx context.Context) ([]models.TableStat, error) {
	rows, err := a.db.QueryContext(ctx, `
		SELECT
			t.TABLE_SCHEMA,
			t.TABLE_NAME,
			IFNULL(p.COUNT_STAR, 0) AS seq_scans,
			t.TABLE_ROWS,
			t.DATA_LENGTH + t.INDEX_LENGTH AS table_size_bytes
		FROM information_schema.TABLES t
		LEFT JOIN performance_schema.table_io_waits_summary_by_index_usage p
			ON p.OBJECT_SCHEMA = t.TABLE_SCHEMA
			AND p.OBJECT_NAME = t.TABLE_NAME
			AND p.INDEX_NAME IS NULL
		WHERE t.TABLE_SCHEMA NOT IN ('mysql', 'performance_schema', 'information_schema', 'sys')
		  AND t.TABLE_TYPE = 'BASE TABLE'
		ORDER BY seq_scans DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("mysql table stats: %w", err)
	}
	defer rows.Close()

	var result []models.TableStat
	for rows.Next() {
		var s models.TableStat
		if err := rows.Scan(&s.SchemaName, &s.TableName, &s.SeqScans, &s.NLiveTup, &s.TableSizeB); err != nil {
			return nil, fmt.Errorf("scan table stat: %w", err)
		}
		result = append(result, s)
	}
	return result, rows.Err()
}
