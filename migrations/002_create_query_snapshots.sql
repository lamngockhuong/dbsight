CREATE TABLE IF NOT EXISTS query_snapshots (
    id            BIGSERIAL PRIMARY KEY,
    connection_id BIGINT NOT NULL REFERENCES connections(id) ON DELETE CASCADE,
    queries       JSONB NOT NULL DEFAULT '[]',
    captured_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_query_snapshots_conn_time
    ON query_snapshots(connection_id, captured_at DESC);
