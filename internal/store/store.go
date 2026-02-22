package store

import (
	"context"

	"github.com/lamngockhuong/dbsight/internal/models"
)

type Store interface {
	// Connections
	CreateConnection(ctx context.Context, c *models.Connection) error
	GetConnection(ctx context.Context, id int64) (*models.Connection, error)
	ListConnections(ctx context.Context) ([]models.Connection, error)
	UpdateConnection(ctx context.Context, c *models.Connection) error
	DeleteConnection(ctx context.Context, id int64) error

	// Query snapshots
	SaveQuerySnapshot(ctx context.Context, snap *models.QuerySnapshot) error
	GetLatestQuerySnapshot(ctx context.Context, connID int64) (*models.QuerySnapshot, error)
	ListQuerySnapshots(ctx context.Context, connID int64, limit int) ([]models.QuerySnapshot, error)

	// Index stats
	SaveIndexStatsSnapshot(ctx context.Context, connID int64, stats []models.IndexStat) error
	GetLatestIndexStats(ctx context.Context, connID int64) ([]models.IndexStat, error)
}
