package adapter

import (
	"context"
	"fmt"

	"github.com/lamngockhuong/dbsight/internal/models"
)

// GetDatabaseStats returns high-level database metrics for MariaDB.
func (a *MariaDBAdapter) GetDatabaseStats(ctx context.Context) (*models.DatabaseStats, error) {
	var s models.DatabaseStats
	var varName string

	if err := a.db.QueryRowContext(ctx, "SELECT DATABASE()").Scan(&s.DBName); err != nil {
		return nil, fmt.Errorf("mariadb db name: %w", err)
	}

	if err := a.db.QueryRowContext(ctx, `
		SELECT IFNULL(SUM(data_length + index_length), 0)
		FROM information_schema.TABLES
		WHERE table_schema = DATABASE()
	`).Scan(&s.SizeBytes); err != nil {
		return nil, fmt.Errorf("mariadb db size: %w", err)
	}

	if err := a.db.QueryRowContext(ctx,
		"SHOW GLOBAL STATUS LIKE 'Threads_connected'",
	).Scan(&varName, &s.ActiveConns); err != nil {
		return nil, fmt.Errorf("mariadb active conns: %w", err)
	}

	if err := a.db.QueryRowContext(ctx,
		"SHOW VARIABLES LIKE 'max_connections'",
	).Scan(&varName, &s.MaxConns); err != nil {
		return nil, fmt.Errorf("mariadb max conns: %w", err)
	}

	var reads, readRequests float64
	if err := a.db.QueryRowContext(ctx,
		"SHOW GLOBAL STATUS LIKE 'Innodb_buffer_pool_reads'",
	).Scan(&varName, &reads); err != nil {
		return nil, fmt.Errorf("mariadb buffer reads: %w", err)
	}
	if err := a.db.QueryRowContext(ctx,
		"SHOW GLOBAL STATUS LIKE 'Innodb_buffer_pool_read_requests'",
	).Scan(&varName, &readRequests); err != nil {
		return nil, fmt.Errorf("mariadb buffer read requests: %w", err)
	}
	if readRequests > 0 {
		s.CacheHitRatio = (1 - reads/readRequests) * 100
	}

	return &s, nil
}
