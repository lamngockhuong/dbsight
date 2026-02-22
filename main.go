package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lamngockhuong/dbsight/internal/adapter"
	"github.com/lamngockhuong/dbsight/internal/api"
	"github.com/lamngockhuong/dbsight/internal/config"
	"github.com/lamngockhuong/dbsight/internal/crypto"
	"github.com/lamngockhuong/dbsight/internal/store"
	"github.com/lamngockhuong/dbsight/internal/worker"
	"github.com/spf13/cobra"
)

//go:embed web/dist
var webDist embed.FS

func main() {
	root := &cobra.Command{
		Use:   "dbsight",
		Short: "Database performance analyzer",
	}
	root.AddCommand(serveCmd())
	root.AddCommand(migrateCmd())
	if err := root.Execute(); err != nil {
		log.Fatal(err)
	}
}

func serveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Start the API server with embedded worker",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Load()
			return runServer(cfg)
		},
	}
}

func migrateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Load()
			ctx := context.Background()
			pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
			if err != nil {
				return fmt.Errorf("connect db: %w", err)
			}
			defer pool.Close()
			return store.RunMigrations(ctx, pool)
		},
	}
}

func runServer(cfg *config.Config) error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	key, err := crypto.KeyFromHex(cfg.EncryptionKey)
	if err != nil {
		return fmt.Errorf("invalid ENCRYPTION_KEY: %w", err)
	}

	st, err := store.NewPGStore(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("store init: %w", err)
	}
	defer st.Close()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("migration pool: %w", err)
	}
	if err := store.RunMigrations(ctx, pool); err != nil {
		pool.Close()
		return fmt.Errorf("migrations: %w", err)
	}
	pool.Close()

	app := &api.App{
		Store:      st,
		CryptoKey:  key,
		NewAdapter: adapter.NewAdapter,
	}

	webFS, err := fs.Sub(webDist, "web/dist")
	if err != nil {
		return fmt.Errorf("static files: %w", err)
	}

	router := api.NewRouter(app, webFS)

	go func() {
		if err := worker.Run(ctx, worker.Config{
			IntervalSecs: cfg.WorkerInterval,
			CryptoKey:    key,
		}, st, adapter.NewAdapter); err != nil && ctx.Err() == nil {
			slog.Error("worker stopped", "err", err)
		}
	}()

	srv := &http.Server{Addr: ":" + cfg.Port, Handler: router}

	go func() {
		<-ctx.Done()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		srv.Shutdown(shutdownCtx)
	}()

	slog.Info("dbsight server starting", "port", cfg.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}
