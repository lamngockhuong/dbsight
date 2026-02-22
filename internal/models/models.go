package models

import (
	"encoding/json"
	"time"
)

type Connection struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	DBType       string    `json:"db_type"`
	EncryptedDSN []byte    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type SlowQuery struct {
	QueryID     string    `json:"query_id"`
	Query       string    `json:"query"`
	Calls       int64     `json:"calls"`
	TotalExecMs float64   `json:"total_exec_ms"`
	MeanExecMs  float64   `json:"mean_exec_ms"`
	Rows        int64     `json:"rows"`
	SnapshotAt  time.Time `json:"snapshot_at"`
}

type QueryDelta struct {
	SlowQuery
	CallsDelta     int64   `json:"calls_delta"`
	TotalExecDelta float64 `json:"total_exec_delta_ms"`
	MeanExecDelta  float64 `json:"mean_exec_delta_ms"`
	PeriodSecs     float64 `json:"period_secs"`
}

type QuerySnapshot struct {
	ID           int64       `json:"id"`
	ConnectionID int64       `json:"connection_id"`
	Queries      []SlowQuery `json:"queries"`
	CapturedAt   time.Time   `json:"captured_at"`
}

type IndexStat struct {
	SchemaName string `json:"schema_name"`
	TableName  string `json:"table_name"`
	IndexName  string `json:"index_name"`
	IndexScans int64  `json:"index_scans"`
	TupRead    int64  `json:"tup_read"`
	TupFetch   int64  `json:"tup_fetch"`
	IndexSizeB int64  `json:"index_size_bytes"`
	IsUnused   bool   `json:"is_unused"`
}

type ExplainPlan struct {
	QueryText string          `json:"query"`
	PlanJSON  json.RawMessage `json:"plan"`
}

type TableStat struct {
	SchemaName string `json:"schema_name"`
	TableName  string `json:"table_name"`
	SeqScans   int64  `json:"seq_scans"`
	SeqTupRead int64  `json:"seq_tup_read"`
	NLiveTup   int64  `json:"n_live_tup"`
	TableSizeB int64  `json:"table_size_bytes"`
}

type DatabaseStats struct {
	DBName        string  `json:"db_name"`
	SizeBytes     int64   `json:"size_bytes"`
	ActiveConns   int     `json:"active_conns"`
	MaxConns      int     `json:"max_conns"`
	CacheHitRatio float64 `json:"cache_hit_ratio"`
}
