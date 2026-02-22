# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

DBSight is a database performance analyzer — Go API server with embedded background worker + React SPA. Single binary serves API, worker (goroutine), and static files. Targets PostgreSQL via `pg_stat_statements`.

## Commands

### Build

```bash
make build                    # Build frontend then Go binary → bin/dbsight
go build -o bin/dbsight .     # Go only (requires web/dist/ from prior pnpm build)
cd web && pnpm run build       # Frontend only (tsc + vite)
```

### Development

```bash
docker-compose up -d postgres  # Start local PostgreSQL
go run . migrate               # Run DB migrations
go run . serve                 # Start API server (port 8080) with embedded worker
cd web && pnpm run dev          # Vite dev server with /api proxy to :8080
```

### Test & Lint

```bash
go test ./internal/...         # Go tests
go vet ./internal/...          # Go static analysis
cd web && pnpm run lint         # Biome (lint + format check)
cd web && pnpm run lint:fix     # Biome auto-fix
cd web && pnpm run format       # Biome format only
```

### Single test file

```bash
go test ./internal/crypto/ -v -run TestEncryptDecrypt
```

## Architecture

**Single binary** (`main.go` at repo root) with cobra CLI:

- `dbsight serve` — starts HTTP server + worker goroutine
- `dbsight migrate` — runs embedded SQL migrations

**`//go:embed web/dist`** in `main.go` embeds the React build into the binary. Must run `pnpm run build` in `web/` before `go build`.

### Backend (Go)

```bash
main.go                          — Entry point, cobra CLI, runServer() wiring, embed directive
internal/config/                 — Config struct loaded from env vars
internal/models/                 — Domain types (Connection, SlowQuery, QueryDelta, etc.)
internal/store/store.go          — Store interface (connections, snapshots, index stats)
internal/store/postgres.go       — pgxpool implementation of Store
internal/store/migrate.go        — Embedded SQL migration runner
internal/adapter/adapter.go      — DBAnalyzer interface + factory (extensible to MySQL etc.)
internal/adapter/postgres.go     — PostgreSQL adapter (Connect/Close)
internal/adapter/slow_queries.go — pg_stat_statements queries
internal/adapter/explain.go      — EXPLAIN plan (read-only tx for safety)
internal/adapter/indexes.go      — Index + table stats from pg_stat_user_indexes
internal/adapter/stats.go        — Database-level stats
internal/api/router.go           — Chi router, CORS, SPA fallback
internal/api/handlers/           — HTTP handlers (connection CRUD, queries, SSE, paste)
internal/worker/scheduler.go     — Ticker-based worker with concurrency limit (10)
internal/worker/collector.go     — Per-connection metrics collector
internal/crypto/encrypt.go       — AES-256-GCM encrypt/decrypt for DSN storage
migrations/                      — SQL files embedded via migrations/embed.go
```

**Key patterns:**

- `App` struct in `internal/api/app.go` holds dependencies (Store, CryptoKey, NewAdapter factory)
- All adapter code lives in `internal/adapter/` package (flat, not subpackaged) to avoid import cycles
- Migrations use `//go:embed` in `migrations/embed.go`, consumed by `store.RunMigrations()`
- Worker starts as `go worker.Run(ctx, ...)` inside `runServer()`

### Frontend (React + TypeScript)

```bash
web/src/
  types/index.ts              — TS interfaces mirroring Go models
  api/client.ts               — Typed fetch wrapper for all API endpoints
  hooks/                      — use-connections, use-queries, use-sse
  components/layout/          — Sidebar + Layout shell
  components/connections/     — ConnectionForm, ConnectionList
  components/queries/         — SlowQueryTable (TanStack Table v8), QueryDetailDrawer, QuerySparkline
  components/ui/              — shadcn/ui components (DO NOT edit manually — use `npx shadcn@latest add`)
  pages/                      — Route pages (kebab-case filenames)
  App.tsx                     — React Router wiring
```

**Key patterns:**

- `@/` path alias configured in tsconfig + vite for imports
- shadcn/ui + Tailwind CSS v4 for styling
- Vite proxy: `/api` → `http://localhost:8080` in dev mode
- SSE hook (`use-sse.ts`) with auto-reconnect for live query updates

### API Endpoints

```bash
GET/POST       /api/connections
GET/PUT/DELETE /api/connections/{id}
POST           /api/connections/{id}/test
GET            /api/connections/{id}/queries          — Latest snapshot with deltas
GET            /api/connections/{id}/queries/stream    — SSE live updates
GET            /api/connections/{id}/queries/history   — Historical snapshots
POST           /api/paste/queries                      — Parse slow log text
/*             — SPA fallback (serves web/dist/index.html)
```

## Environment Variables

| Variable               | Default    | Description                                            |
| ---------------------- | ---------- | ------------------------------------------------------ |
| `PORT`                 | `8080`     | HTTP server port                                       |
| `DATABASE_URL`         | (required) | PostgreSQL connection string for app metadata DB       |
| `ENCRYPTION_KEY`       | (required) | 64 hex chars (32 bytes) for AES-256-GCM DSN encryption |
| `WORKER_INTERVAL_SECS` | `30`       | Background worker polling interval                     |

## Security Notes

- DSN stored as AES-256-GCM ciphertext in `connections.encrypted_dsn` (BYTEA)
- `EncryptedDSN` field has `json:"-"` tag — never serialized to API responses
- EXPLAIN uses `SET TRANSACTION READ ONLY` to mitigate injection risk
- All store queries use parameterized placeholders
- Request body limits: 1MB for CRUD, 10MB for paste endpoint
- No authentication in MVP — localhost-only tool
