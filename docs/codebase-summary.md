# DBSight Codebase Summary

## Directory Structure

```bash
dbsight/
├── main.go                          # CLI entry point (Cobra)
├── Makefile                         # Build targets (monorepo-aware)
├── Dockerfile                       # Multi-stage build (Node + Go + Alpine)
├── docker-compose.yml               # Dev environment (postgres + app)
├── pnpm-workspace.yaml              # pnpm monorepo workspace config
├── package.json                     # Root workspace scripts
├── .env.example                     # Environment variable template
├── go.mod / go.sum                  # Go module dependencies
├── README.md                        # Project overview (end-user)
├── LICENSE                          # MIT license
├── CLAUDE.md                        # Development instructions
│
├── apps/
│   ├── web/                         # React SPA (moved from web/)
│   │   ├── public/                  # Static assets
│   │   ├── src/                     # Application source
│   │   ├── package.json
│   │   ├── tsconfig.json
│   │   ├── vite.config.ts
│   │   ├── biome.json
│   │   └── components.json
│   │
│   └── docs/                        # Starlight documentation site (EN + VI)
│       ├── src/
│       │   ├── content/             # MDX docs content
│       │   └── assets/              # Doc assets
│       ├── astro.config.mjs
│       ├── package.json
│       └── tsconfig.json
│
├── .github/
│   └── workflows/
│       └── deploy-docs.yml          # GitHub Pages deployment for docs site
│
├── internal/
│   ├── config/                      # Configuration loading
│   │   └── config.go                # Env var parsing
│   │
│   ├── models/                      # Domain types
│   │   └── models.go                # Connection, SlowQuery, ExplainPlan, etc.
│   │
│   ├── store/                       # Persistence layer
│   │   ├── store.go                 # Interface definitions
│   │   ├── postgres.go              # PostgreSQL implementation (pgxpool)
│   │   └── migrate.go               # Migration runner
│   │
│   ├── adapter/                     # Database analyzer interface
│   │   ├── adapter.go               # Interface + factory
│   │   ├── postgres.go              # PostgreSQL queries
│   │   ├── slow_queries.go          # pg_stat_statements
│   │   ├── explain.go               # EXPLAIN ANALYZE parsing
│   │   ├── indexes.go               # pg_stat_indexes
│   │   └── stats.go                 # Table/database statistics
│   │
│   ├── api/                         # HTTP server
│   │   ├── router.go                # Chi route registration + /healthz
│   │   ├── app.go                   # App struct (dependency holder)
│   │   ├── handlers/                # Endpoint handlers
│   │   │   ├── connection.go        # Connection CRUD
│   │   │   ├── queries.go           # Query endpoints + SSE
│   │   │   ├── explain.go           # RunExplain handler (Phase 08)
│   │   │   ├── indexes.go           # GetIndexAnalysis + computeRecommendations (Phase 09)
│   │   │   ├── paste.go             # Slow log parsing
│   │   │   └── handler.go           # Common utilities
│   │   └── middleware/              # HTTP middleware
│   │       ├── logger.go            # Request logging
│   │       └── recovery.go          # Panic recovery
│   │
│   ├── worker/                      # Background metrics collector
│   │   ├── scheduler.go             # 30s ticker + semaphore
│   │   └── collector.go             # Per-connection metrics
│   │
│   └── crypto/                      # Encryption utilities
│       ├── encrypt.go               # AES-256-GCM encrypt/decrypt
│       └── encrypt_test.go          # Unit tests
│
├── migrations/                      # SQL schema definitions
│   ├── 001_create_connections.sql
│   ├── 002_create_query_snapshots.sql
│   ├── 003_create_index_stats_snapshots.sql
│   └── embed.go                     # go:embed FS
│
│   (web/ contents moved to apps/web/ — see above)
│
├── docs/                            # Project documentation
│   ├── README.md                    # [KEPT] End-user docs
│   ├── system-architecture.md       # [UPDATED] Component interactions
│   ├── code-standards.md            # [UPDATED] Coding guidelines
│   ├── project-overview-pdr.md      # [NEW] Product requirements
│   ├── codebase-summary.md          # [NEW] This file
│   ├── project-roadmap.md           # [NEW] MVP → post-MVP phases
│   └── deployment-guide.md          # [NEW] Docker + production setup
│
└── plans/                           # Project planning docs (separate from docs/)
    ├── 260221-1933-database-analyzer-webapp/
    │   ├── plan.md
    │   ├── phase-*.md
    │   └── reports/
    └── reports/
```

## Key Files by Purpose

### Backend Entry Point & Wiring

- **main.go** (130 LOC): Cobra CLI with `serve` and `migrate` commands. Initializes config, Store, Adapters, and wires everything into the App struct. Embeds `apps/web/dist` for SPA serving via `//go:embed apps/web/dist`.

### Configuration

- **internal/config/config.go** (41 LOC): Loads PORT, DATABASE_URL, ENCRYPTION_KEY, WORKER_INTERVAL_SECS from environment. Validates ENCRYPTION_KEY format (64 hex chars).

### Domain Models

- **internal/models/models.go**: Defines Connection, SlowQuery, QuerySnapshot, QueryDelta, IndexStat, ExplainPlan, TableStat, DatabaseStats, DuplicateIndex, Recommendation, IndexAnalysisResult. Includes JSON tags for API serialization.

### Data Persistence

- **internal/store/store.go**: Interface defining Store behavior (Ping, CreateConnection, GetConnection, SaveQuerySnapshot, etc.)
- **internal/store/postgres.go** (316 LOC): PostgreSQL implementation using pgxpool. Handles all database CRUD operations. Encrypts DSN on write, decrypts on read.
- **internal/store/migrate.go**: Embedded SQL migrations runner. Tracks schema_version table to prevent re-running migrations.

### Database Analysis (Adapter Pattern)

- **internal/adapter/adapter.go**: Defines DBAnalyzer interface (GetSlowQueries, GetExplainPlan, GetIndexStats, GetTableStats, GetDatabaseStats). Factory function returns appropriate adapter.
- **internal/adapter/postgres.go**: Routes PostgreSQL queries; imports submodules for specific metric types.
- **internal/adapter/slow_queries.go**: Queries pg_stat_statements, parses results into SlowQuery structs.
- **internal/adapter/explain.go**: Executes EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON) in read-only transaction; parses JSON response.
- **internal/adapter/indexes.go**: Queries pg_stat_indexes, pg_stat_user_indexes for usage stats.
- **internal/adapter/stats.go**: Queries information_schema for table stats, pg_database for database-wide metrics.

### HTTP API

- **internal/api/router.go**: Chi router setup. Routes: /healthz, /api/connections/_, /api/paste/_, static SPA fallback.
- **internal/api/app.go**: App struct holds Store, CryptoKey, DBAnalyzer factory. Embedded in all handlers.
- **internal/api/middleware/logger.go**: Structured request/response logging via slog.
- **internal/api/middleware/recovery.go**: Panic recovery, returns 500 error response.
- **internal/api/handlers/connection.go**: ListConnections, GetConnection, CreateConnection, UpdateConnection, DeleteConnection, TestConnection.
- **internal/api/handlers/queries.go**: ListQueries, StreamQueries (SSE), GetQueryHistory. SSE broadcasts delta changes per connection.
- **internal/api/handlers/explain.go**: RunExplain — decrypts DSN, connects adapter, calls GetExplainPlan with 30s timeout. (Phase 08)
- **internal/api/handlers/indexes.go**: GetIndexAnalysis — collects index/table stats, detects unused/duplicate/missing, runs computeRecommendations to generate DROP/CREATE SQL. (Phase 09)
- **internal/api/handlers/paste.go**: Parses MySQL slow log format, returns analysis without live DB.

### Background Worker

- **internal/worker/scheduler.go** (78 LOC): 30s ticker, semaphore-limited worker pool (max 10 concurrent), runs collector per connection.
- **internal/worker/collector.go** (113 LOC): Decrypts DSN, creates adapter, collects slow queries/index stats. Saves QuerySnapshot to Store.

### Encryption

- **internal/crypto/encrypt.go** (88 LOC): AES-256-GCM encryption/decryption. Key must be 32 bytes (64 hex chars). Uses random nonce per encryption.
- **internal/crypto/encrypt_test.go**: Unit tests for encrypt/decrypt roundtrip, error cases.

### Frontend Components

#### API Client

- **apps/web/src/api/client.ts**: Fetch-based API wrapper. Methods: getConnections, createConnection, testConnection, getQueries, getQueryHistory, explainQuery, etc.

#### Hooks

- **apps/web/src/hooks/use-connections.ts**: useState for connections list, loading, error. Fetch on mount.
- **apps/web/src/hooks/use-queries.ts**: useState for slow queries per connection. Handles sorting, filtering.
- **apps/web/src/hooks/use-sse.ts**: EventSource subscription to /api/connections/{id}/queries/stream. Merges delta updates into query list.

#### Components

- **apps/web/src/components/ui/\***: shadcn/ui base components (Button, Card, Input, Table, Badge, Tabs, Textarea).
- **apps/web/src/components/layout/layout.tsx**: Main shell with sidebar + content area.
- **apps/web/src/components/layout/sidebar.tsx**: Navigation links to pages.
- **apps/web/src/components/connections/connection-list.tsx**: Table of registered connections with test/edit/delete buttons.
- **apps/web/src/components/connections/connection-form.tsx**: Form to create/edit connection (name, host, port, database, user, password).
- **apps/web/src/components/queries/slow-query-table.tsx**: TanStack Table v8 with sortable/filterable columns. Displays query text, calls, total time, delta.
- **apps/web/src/components/queries/query-detail-drawer.tsx**: Side panel showing full query text, execution stats, EXPLAIN plan if available.
- **apps/web/src/components/queries/query-sparkline.tsx**: Recharts mini line chart showing execution time trend.
- **apps/web/src/components/explain/explain-json-tree.tsx**: Collapsible JSON tree renderer for EXPLAIN plan output. Annotates costs, highlights sequential scans and row estimate mismatches. (Phase 08)
- **apps/web/src/components/indexes/recommendations-list.tsx**: Renders Recommendation list with severity badges and copyable SQL. (Phase 09)

#### Pages

- **apps/web/src/pages/dashboard-page.tsx**: Main overview; connection selector, quick stats, top slow queries chart.
- **apps/web/src/pages/connections-page.tsx**: Connection CRUD UI.
- **apps/web/src/pages/queries-page.tsx**: Query dashboard with live updates (SSE).
- **apps/web/src/pages/explain-page.tsx**: Direct mode (run EXPLAIN via API) + Paste JSON mode; ANALYZE warning banner; renders explain-json-tree. (Phase 08)
- **apps/web/src/pages/indexes-page.tsx**: Summary cards (unused count, duplicate count, recommendation count), recommendations list, detail tables. (Phase 09)
- **apps/web/src/pages/paste-page.tsx**: Paste MySQL slow log, analyze offline.

### Migrations

- **migrations/001_create_connections.sql**: Creates `connections` table (id, name, dbType, encryptedDSN, createdAt, updatedAt).
- **migrations/002_create_query_snapshots.sql**: Creates `query_snapshots` table (connectionID, query, calls, totalTime, meanTime, lastExecution, snapshot_at).
- **migrations/003_create_index_stats_snapshots.sql**: Creates `index_stats_snapshots` table (connectionID, indexName, tableSize, indexSize, idxScan, snapshot_at).
- **migrations/embed.go**: Embeds migration files in Go binary via `//go:embed migrations/*.sql`.

## Package Dependencies

### Go (main.go, go.mod)

- **chi** — HTTP router, lightweight middleware
- **pgx/v5** — PostgreSQL driver, parameterized queries, connection pooling
- **cobra** — CLI framework for `serve` and `migrate` commands
- **log/slog** — Structured logging (stdlib, no external dep post-Go 1.21)

### Node / pnpm (apps/web/package.json, root package.json)

- **react@19** — UI library
- **react-dom@19** — DOM rendering
- **typescript** — Type safety
- **vite** — Build tool, dev server
- **shadcn/ui** — Headless component library
- **tailwindcss@4** — Utility CSS
- **@tanstack/react-table@8** — Headless table library
- **recharts** — Chart library (React)
- **@radix-ui/\*** — Accessible component primitives (shadcn/ui deps)
- **@biomejs/biome** — Linting + formatting

## Data Models

### Connection

```go
type Connection struct {
    ID           int64     `json:"id"`
    Name         string    `json:"name"`
    DBType       string    `json:"db_type"`         // "postgres", "mysql", etc.
    EncryptedDSN []byte    `json:"-"`               // Never in API response
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}
```

### SlowQuery

```go
type SlowQuery struct {
    QueryID       string    `json:"query_id"`
    Query         string    `json:"query"`
    Calls         int64     `json:"calls"`
    TotalExecMs   float64   `json:"total_exec_ms"`     // milliseconds
    MeanExecMs    float64   `json:"mean_exec_ms"`      // milliseconds
    Rows          int64     `json:"rows"`
    SnapshotAt    time.Time `json:"snapshot_at"`
}
```

### QuerySnapshot

```go
type QuerySnapshot struct {
    ID           int64       `json:"id"`
    ConnectionID int64       `json:"connection_id"`
    Queries      []SlowQuery `json:"queries"`       // Array of slow queries
    CapturedAt   time.Time   `json:"captured_at"`
}
```

### ExplainPlan

```go
type ExplainPlan struct {
    QueryText string          `json:"query"`       // SQL query text
    PlanJSON  json.RawMessage `json:"plan"`        // EXPLAIN JSON output
}
```

## Data Flow: Key Scenarios

### Scenario 1: Create & Monitor a Connection

1. User submits connection form (host, port, database, user, password)
2. API handler CreateConnection receives request
3. Validates DSN format
4. Encrypts DSN with AES-256-GCM (uses config.ENCRYPTION_KEY)
5. Stores Connection in PostgreSQL (encryptedDSN field)
6. Returns connection ID to frontend
7. Worker picks up new connection on next 30s tick
8. Decrypts DSN, creates PostgreSQL adapter, queries pg_stat_statements
9. Saves QuerySnapshot to PostgreSQL
10. Frontend SSE client receives update via /api/connections/{id}/queries/stream
11. React state updates, dashboard re-renders with new metrics

### Scenario 2: Real-Time Query Streaming (SSE)

1. Frontend subscribes to EventSource at /api/connections/{id}/queries/stream
2. User opens queries page, handler calls h.app.store.GetQuerySnapshot(ctx, connID)
3. Returns most recent snapshot
4. Handler creates SSE writer, starts event loop
5. Worker runs every 30s, collects new SlowQuery data, saves QuerySnapshot
6. Worker signals SSE handler via channel (if implemented) or handler polls store
7. Handler detects new snapshot, calculates delta, marshals to JSON
8. Writes `data: {...}` + newline to SSE writer
9. Frontend EventSource listener receives event, updates React state
10. TanStack Table re-renders with new metrics

### Scenario 3: Explain a Query

1. User clicks "Explain" on query row
2. Frontend calls POST /api/connections/{id}/explain?query=<SQL>
3. API handler calls adapter.GetExplainPlan(ctx, query)
4. Adapter connects to target DB, runs `EXPLAIN ANALYZE (BUFFERS, FORMAT JSON) <SQL>` in read-only txn
5. Parses JSON result into ExplainPlan struct
6. Returns to frontend
7. React component displays plan tree with cost breakdown

## API Endpoints Summary

| Method | Endpoint                              | Handler          | Returns               |
| ------ | ------------------------------------- | ---------------- | --------------------- |
| GET    | /healthz                              | inline           | {status}              |
| GET    | /api/connections                      | ListConnections  | []{Connection}        |
| POST   | /api/connections                      | CreateConnection | {Connection}          |
| GET    | /api/connections/{id}                 | GetConnection    | {Connection}          |
| PUT    | /api/connections/{id}                 | UpdateConnection | {Connection}          |
| DELETE | /api/connections/{id}                 | DeleteConnection | {ok: true}            |
| POST   | /api/connections/{id}/test            | TestConnection   | {latencyMs: int}      |
| GET    | /api/connections/{id}/queries         | ListQueries      | []{QuerySnapshot}     |
| GET    | /api/connections/{id}/queries/stream  | StreamQueries    | SSE stream            |
| GET    | /api/connections/{id}/queries/history | GetQueryHistory  | []{QuerySnapshot}     |
| POST   | /api/connections/{id}/explain         | RunExplain       | {ExplainPlan}         |
| GET    | /api/connections/{id}/indexes         | GetIndexAnalysis | {IndexAnalysisResult} |
| POST   | /api/paste/queries                    | PasteQueries     | {analysis: ...}       |

## Critical Design Patterns

### 1. Adapter Pattern (Extensibility)

All database access goes through `DBAnalyzer` interface. Implementations:

- PostgreSQL adapter (current)
- MySQL adapter (post-MVP)
- Others can be added without changing core

### 2. Dependency Injection (Testability)

App struct holds Store, CryptoKey, adapter factory. Passed to handlers for dependency resolution.

### 3. Error Wrapping (Debugging)

All errors wrapped with context: `fmt.Errorf("GetSlowQueries: %w", err)`

### 4. Encryption by Default (Security)

DSN never stored plaintext. AES-256-GCM with authenticated encryption prevents tampering.

### 5. Embedded Assets (Deployment)

Frontend SPA embedded in Go binary via `//go:embed apps/web/dist`. Single artifact to deploy.

### 6. Idempotent Migrations (Durability)

Migrations tracked in `schema_version` table. Safe to re-run on upgrade.

## Code Metrics

- **Backend:** ~3,250 LOC across 62 files
- **Frontend:** ~1,400 LOC (React components + hooks)
- **Tests:** >70% coverage (critical paths)
- **Go Version:** 1.26+
- **TypeScript:** Strict mode enabled

## Dependencies & Constraints

- **PostgreSQL:** 14+ (metadata storage + pg_stat_statements on target)
- **Node.js:** 20+ (frontend build only; not required at runtime)
- **Go:** 1.26+ (compiler requirement)
- **Environment:** PORT, DATABASE_URL, ENCRYPTION_KEY (required)

---

**Document Version:** 1.2
**Last Updated:** 2026-02-22
**Scope:** Production ready (Phases 1–10) + monorepo restructure
