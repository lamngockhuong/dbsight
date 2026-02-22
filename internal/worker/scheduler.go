package worker

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/lamngockhuong/dbsight/internal/adapter"
	"github.com/lamngockhuong/dbsight/internal/models"
	"github.com/lamngockhuong/dbsight/internal/store"
)

const maxConcurrentCollectors = 10

type Config struct {
	IntervalSecs int
	CryptoKey    []byte
}

func Run(ctx context.Context, cfg Config, st store.Store,
	newAdapter func(string) (adapter.DBAnalyzer, error)) error {

	ticker := time.NewTicker(time.Duration(cfg.IntervalSecs) * time.Second)
	defer ticker.Stop()
	slog.Info("worker started", "interval_secs", cfg.IntervalSecs)

	for {
		select {
		case <-ctx.Done():
			slog.Info("worker stopping")
			return ctx.Err()
		case <-ticker.C:
			conns, err := st.ListConnections(ctx)
			if err != nil {
				slog.Error("worker list connections", "err", err)
				continue
			}
			sem := make(chan struct{}, maxConcurrentCollectors)
			var wg sync.WaitGroup
			for _, conn := range conns {
				wg.Add(1)
				sem <- struct{}{}
				go func(c models.Connection) {
					defer wg.Done()
					defer func() { <-sem }()
					CollectMetrics(ctx, c, cfg.CryptoKey, st, newAdapter)
				}(conn)
			}
			wg.Wait()
		}
	}
}
