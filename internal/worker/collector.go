package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/lamngockhuong/dbsight/internal/adapter"
	"github.com/lamngockhuong/dbsight/internal/crypto"
	"github.com/lamngockhuong/dbsight/internal/models"
	"github.com/lamngockhuong/dbsight/internal/store"
)

func CollectMetrics(ctx context.Context, conn models.Connection, cryptoKey []byte,
	st store.Store, newAdapter func(string) (adapter.DBAnalyzer, error)) {

	dsn, err := crypto.Decrypt(cryptoKey, conn.EncryptedDSN)
	if err != nil {
		slog.Error("decrypt dsn", "conn_id", conn.ID, "err", err)
		return
	}

	a, err := newAdapter(conn.DBType)
	if err != nil {
		slog.Error("new adapter", "conn_id", conn.ID, "err", err)
		return
	}
	defer a.Close()

	connCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := a.Connect(connCtx, string(dsn)); err != nil {
		slog.Error("collector connect", "conn_id", conn.ID, "err", err)
		return
	}

	queries, err := a.GetSlowQueries(ctx, adapter.QueryOpts{Limit: 100})
	if err != nil {
		slog.Error("get slow queries", "conn_id", conn.ID, "err", err)
		return
	}

	snap := &models.QuerySnapshot{
		ConnectionID: conn.ID,
		Queries:      queries,
		CapturedAt:   time.Now(),
	}
	if err := st.SaveQuerySnapshot(ctx, snap); err != nil {
		slog.Error("save snapshot", "conn_id", conn.ID, "err", err)
	}

	idxStats, err := a.GetIndexStats(ctx)
	if err != nil {
		slog.Warn("get index stats", "conn_id", conn.ID, "err", err)
		return
	}
	if err := st.SaveIndexStatsSnapshot(ctx, conn.ID, idxStats); err != nil {
		slog.Error("save index stats", "conn_id", conn.ID, "err", err)
	}
}
