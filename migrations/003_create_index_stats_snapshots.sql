CREATE TABLE IF NOT EXISTS index_stats_snapshots (
    id            BIGSERIAL PRIMARY KEY,
    connection_id BIGINT NOT NULL REFERENCES connections(id) ON DELETE CASCADE,
    stats         JSONB NOT NULL DEFAULT '[]',
    captured_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_index_stats_conn_time
    ON index_stats_snapshots(connection_id, captured_at DESC);
