package adapter

import (
	"context"
	"fmt"

	"github.com/lamngockhuong/dbsight/internal/models"
)

// QueryOpts configures query collection behaviour.
type QueryOpts struct {
	Limit       int
	MinMeanMs   float64
	AnalyzeMode bool
}

// DBAnalyzer is the interface every database adapter must satisfy.
type DBAnalyzer interface {
	Connect(ctx context.Context, dsn string) error
	Close() error
	GetSlowQueries(ctx context.Context, opts QueryOpts) ([]models.SlowQuery, error)
	GetExplainPlan(ctx context.Context, query string, opts QueryOpts) (*models.ExplainPlan, error)
	GetIndexStats(ctx context.Context) ([]models.IndexStat, error)
	GetTableStats(ctx context.Context) ([]models.TableStat, error)
	GetDatabaseStats(ctx context.Context) (*models.DatabaseStats, error)
}

// NewAdapter returns a DBAnalyzer for the given database type.
func NewAdapter(dbType string) (DBAnalyzer, error) {
	switch dbType {
	case "postgres":
		return NewPostgresAdapter(), nil
	default:
		return nil, fmt.Errorf("unsupported db type: %s", dbType)
	}
}
