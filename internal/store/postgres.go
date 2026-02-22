package store

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lamngockhuong/dbsight/internal/models"
)

type PGStore struct {
	pool *pgxpool.Pool
}

func NewPGStore(ctx context.Context, dsn string) (*PGStore, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}
	return &PGStore{pool: pool}, nil
}

func (s *PGStore) Close() {
	s.pool.Close()
}

func (s *PGStore) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

// CreateConnection inserts a new connection and populates id, created_at, updated_at.
func (s *PGStore) CreateConnection(ctx context.Context, c *models.Connection) error {
	return s.pool.QueryRow(ctx, `
		INSERT INTO connections (name, db_type, encrypted_dsn)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`, c.Name, c.DBType, c.EncryptedDSN).Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)
}

// GetConnection retrieves a single connection by id.
func (s *PGStore) GetConnection(ctx context.Context, id int64) (*models.Connection, error) {
	c := &models.Connection{}
	err := s.pool.QueryRow(ctx, `
		SELECT id, name, db_type, encrypted_dsn, created_at, updated_at
		FROM connections WHERE id = $1
	`, id).Scan(&c.ID, &c.Name, &c.DBType, &c.EncryptedDSN, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get connection %d: %w", id, err)
	}
	return c, nil
}

// ListConnections returns all connections ordered by id.
func (s *PGStore) ListConnections(ctx context.Context) ([]models.Connection, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, db_type, encrypted_dsn, created_at, updated_at
		FROM connections ORDER BY id
	`)
	if err != nil {
		return nil, fmt.Errorf("list connections: %w", err)
	}
	defer rows.Close()

	var result []models.Connection
	for rows.Next() {
		var c models.Connection
		if err := rows.Scan(&c.ID, &c.Name, &c.DBType, &c.EncryptedDSN, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan connection: %w", err)
		}
		result = append(result, c)
	}
	return result, rows.Err()
}

// UpdateConnection updates name, db_type, encrypted_dsn and refreshes updated_at.
func (s *PGStore) UpdateConnection(ctx context.Context, c *models.Connection) error {
	return s.pool.QueryRow(ctx, `
		UPDATE connections
		SET name = $1, db_type = $2, encrypted_dsn = $3, updated_at = NOW()
		WHERE id = $4
		RETURNING updated_at
	`, c.Name, c.DBType, c.EncryptedDSN, c.ID).Scan(&c.UpdatedAt)
}

// DeleteConnection removes a connection by id.
func (s *PGStore) DeleteConnection(ctx context.Context, id int64) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM connections WHERE id = $1`, id)
	return err
}

// SaveQuerySnapshot marshals queries to JSONB and inserts.
func (s *PGStore) SaveQuerySnapshot(ctx context.Context, snap *models.QuerySnapshot) error {
	data, err := json.Marshal(snap.Queries)
	if err != nil {
		return fmt.Errorf("marshal queries: %w", err)
	}
	return s.pool.QueryRow(ctx, `
		INSERT INTO query_snapshots (connection_id, queries)
		VALUES ($1, $2)
		RETURNING id, captured_at
	`, snap.ConnectionID, data).Scan(&snap.ID, &snap.CapturedAt)
}

// GetLatestQuerySnapshot returns the most recent snapshot for a connection.
func (s *PGStore) GetLatestQuerySnapshot(ctx context.Context, connID int64) (*models.QuerySnapshot, error) {
	snap := &models.QuerySnapshot{ConnectionID: connID}
	var raw []byte
	err := s.pool.QueryRow(ctx, `
		SELECT id, queries, captured_at
		FROM query_snapshots
		WHERE connection_id = $1
		ORDER BY captured_at DESC LIMIT 1
	`, connID).Scan(&snap.ID, &raw, &snap.CapturedAt)
	if err != nil {
		return nil, fmt.Errorf("get latest snapshot conn %d: %w", connID, err)
	}
	if err := json.Unmarshal(raw, &snap.Queries); err != nil {
		return nil, fmt.Errorf("unmarshal queries: %w", err)
	}
	return snap, nil
}

// ListQuerySnapshots returns up to limit snapshots ordered newest first.
func (s *PGStore) ListQuerySnapshots(ctx context.Context, connID int64, limit int) ([]models.QuerySnapshot, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, queries, captured_at
		FROM query_snapshots
		WHERE connection_id = $1
		ORDER BY captured_at DESC LIMIT $2
	`, connID, limit)
	if err != nil {
		return nil, fmt.Errorf("list snapshots conn %d: %w", connID, err)
	}
	defer rows.Close()

	var result []models.QuerySnapshot
	for rows.Next() {
		snap := models.QuerySnapshot{ConnectionID: connID}
		var raw []byte
		if err := rows.Scan(&snap.ID, &raw, &snap.CapturedAt); err != nil {
			return nil, fmt.Errorf("scan snapshot: %w", err)
		}
		if err := json.Unmarshal(raw, &snap.Queries); err != nil {
			return nil, fmt.Errorf("unmarshal queries: %w", err)
		}
		result = append(result, snap)
	}
	return result, rows.Err()
}

// SaveIndexStatsSnapshot marshals stats and inserts a new snapshot row.
func (s *PGStore) SaveIndexStatsSnapshot(ctx context.Context, connID int64, stats []models.IndexStat) error {
	data, err := json.Marshal(stats)
	if err != nil {
		return fmt.Errorf("marshal index stats: %w", err)
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO index_stats_snapshots (connection_id, stats)
		VALUES ($1, $2)
	`, connID, data)
	return err
}

// GetLatestIndexStats returns the most recent index stats for a connection.
func (s *PGStore) GetLatestIndexStats(ctx context.Context, connID int64) ([]models.IndexStat, error) {
	var raw []byte
	err := s.pool.QueryRow(ctx, `
		SELECT stats
		FROM index_stats_snapshots
		WHERE connection_id = $1
		ORDER BY captured_at DESC LIMIT 1
	`, connID).Scan(&raw)
	if err != nil {
		return nil, fmt.Errorf("get latest index stats conn %d: %w", connID, err)
	}
	var stats []models.IndexStat
	if err := json.Unmarshal(raw, &stats); err != nil {
		return nil, fmt.Errorf("unmarshal index stats: %w", err)
	}
	return stats, nil
}
