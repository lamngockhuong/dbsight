package adapter

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/lamngockhuong/dbsight/internal/adapter/mysqlcompat"
)

// MySQLAdapter implements DBAnalyzer for MySQL 5.7+/8.0+.
type MySQLAdapter struct {
	db      *sql.DB
	version mysqlcompat.Version
}

// NewMySQLAdapter creates an unconnected MySQLAdapter.
func NewMySQLAdapter() *MySQLAdapter { return &MySQLAdapter{} }

// Connect opens a connection to MySQL and verifies connectivity + performance_schema.
func (a *MySQLAdapter) Connect(ctx context.Context, dsn string) error {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("open mysql: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return fmt.Errorf("ping mysql: %w", err)
	}
	v, err := mysqlcompat.DetectVersion(ctx, db)
	if err != nil {
		db.Close()
		return fmt.Errorf("detect mysql version: %w", err)
	}
	enabled, err := mysqlcompat.CheckPerfSchema(ctx, db)
	if err != nil {
		db.Close()
		return fmt.Errorf("check perf_schema: %w", err)
	}
	if !enabled {
		db.Close()
		return fmt.Errorf("performance_schema is disabled; enable it in MySQL config to use DBSight")
	}
	a.db = db
	a.version = v
	return nil
}

// Close releases the database connection.
func (a *MySQLAdapter) Close() error {
	if a.db != nil {
		return a.db.Close()
	}
	return nil
}
