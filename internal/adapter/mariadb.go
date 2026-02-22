package adapter

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/lamngockhuong/dbsight/internal/adapter/mysqlcompat"
)

// MariaDBAdapter implements DBAnalyzer for MariaDB 10.x+.
type MariaDBAdapter struct {
	db      *sql.DB
	version mysqlcompat.Version
}

// NewMariaDBAdapter creates an unconnected MariaDBAdapter.
func NewMariaDBAdapter() *MariaDBAdapter { return &MariaDBAdapter{} }

// Connect opens a connection to MariaDB and verifies connectivity + performance_schema.
func (a *MariaDBAdapter) Connect(ctx context.Context, dsn string) error {
	db, err := sql.Open("mysql", dsn) // same driver as MySQL
	if err != nil {
		return fmt.Errorf("open mariadb: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return fmt.Errorf("ping mariadb: %w", err)
	}
	v, err := mysqlcompat.DetectVersion(ctx, db)
	if err != nil {
		db.Close()
		return fmt.Errorf("detect mariadb version: %w", err)
	}
	enabled, err := mysqlcompat.CheckPerfSchema(ctx, db)
	if err != nil {
		db.Close()
		return fmt.Errorf("check perf_schema: %w", err)
	}
	if !enabled {
		db.Close()
		return fmt.Errorf("performance_schema is disabled; enable it in MariaDB config to use DBSight")
	}
	a.db = db
	a.version = v
	return nil
}

// Close releases the database connection.
func (a *MariaDBAdapter) Close() error {
	if a.db != nil {
		return a.db.Close()
	}
	return nil
}
