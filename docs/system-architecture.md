# System Architecture

## Overview

DBSight is a three-tier application:

1. **Frontend**: React SPA serving an interactive dashboard
2. **Backend API**: Go HTTP server handling connections, queries, and metrics
3. **Data Layer**: PostgreSQL storing metadata and metrics snapshots

The backend runs a background metrics collector (worker goroutine) that polls connected databases via adapters.

## Architecture Diagram

```bash
┌─────────────────────────────────────────────────────────────┐
│                    React SPA (Vite)                         │
│  ┌────────────┬──────────────────┬─────────────────────┐    │
│  │  Dashboard │  Connections     │  Query Analysis     │    │
│  │  (Charts)  │  (Forms)         │  (Detail View)      │    │
│  └────────────┴──────────────────┴─────────────────────┘    │
└────────────────────────┬────────────────────────────────────┘
                         │ HTTP/SSE
                         ▼
┌─────────────────────────────────────────────────────────────┐
│               Go HTTP Server (Chi Router)                   │
│  ┌─────────────┐  ┌──────────────┐  ┌─────────────────┐     │
│  │  API Routes │  │  Middleware  │  │  Static Files   │     │
│  │  (Handlers) │  │  (Logger,    │  │  (SPA)          │     │
│  │             │  │   Recovery)  │  │                 │     │
│  └─────────────┘  └──────────────┘  └─────────────────┘     │
└────────────────────────┬────────────────────────────────────┘
         ┌───────────────┼───────────────┐
         │               │               │
         ▼               ▼               ▼
┌──────────────┐ ┌──────────────┐ ┌─────────────────┐
│  Store       │ │  Adapters    │ │  Crypto Module  │
│  (PostgreSQL)│ │  (Multi-DB)  │ │  (AES-256-GCM)  │
│              │ │              │ │                 │
│ Connections  │ │  PostgreSQL  │ │ Encrypt/Decrypt │
│ Query Logs   │ │  Adapter     │ │ DSN             │
│ Index Stats  │ │              │ │                 │
│ Metrics      │ │  (Extensible)│ │                 │
└──────────────┘ └──────────────┘ └─────────────────┘
         │               │
         └───────────────┼───────────────┐
                         │               │
                         ▼               ▼
                   ┌──────────────────────────────┐
                   │ Target Databases             │
                   │ (PostgreSQL instances)       │
                   │                              │
                   │ pg_stat_statements           │
                   │ pg_stat_indexes              │
                   │ Information Schema           │
                   └──────────────────────────────┘
                         ▲
                         │ Worker Goroutine
                         │ (Polls every 30s)
```

## Core Components

### 1. Backend (Go)

#### Entry Point (`main.go`)

- Initializes Cobra CLI with `serve` and `migrate` commands
- Wires dependencies: Store, Adapters, Crypto module
- Starts HTTP server on configured PORT
- Launches background worker goroutine

#### Configuration (`internal/config`)

Loads environment variables into a Config struct:

- `PORT`: HTTP server port (default: 8080)
- `DATABASE_URL`: PostgreSQL connection string
- `ENCRYPTION_KEY`: 32-byte hex string for AES-256-GCM
- `WORKER_INTERVAL_SECS`: Metrics polling interval (default: 30)

#### Models (`internal/models`)

Domain types representing database entities:

- `Connection`: Registered database connection (id, name, dbType, encryptedDSN, createdAt)
- `SlowQuery`: Query from pg_stat_statements (query, calls, totalTime, meanTime, maxTime)
- `ExplainPlan`: Query execution plan with cost estimation
- `IndexStat`: Index usage statistics (indexName, tableSize, indexSize, idxScan)
- `TableStat`: Table statistics (tableName, rowCount, size, lastVacuum)
- `DatabaseStats`: Database-wide metrics (totalSize, connections, txnStats)

#### Store Layer (`internal/store`)

Interface-driven data persistence:

```go
type Store interface {
    // Connections
    CreateConnection(ctx context.Context, conn *Connection) error
    GetConnection(ctx context.Context, id string) (*Connection, error)
    ListConnections(ctx context.Context) ([]*Connection, error)
    UpdateConnection(ctx context.Context, id string, conn *Connection) error
    DeleteConnection(ctx context.Context, id string) error

    // Query snapshots
    SaveQuerySnapshot(ctx context.Context, snapshot *QuerySnapshot) error
    GetQueryHistory(ctx context.Context, connID string, limit int) ([]*QuerySnapshot, error)
}
```

**PostgreSQL Implementation** (`internal/store`):

- Uses pgxpool for connection pooling
- Parameterized queries to prevent SQL injection
- Automatic schema migration on startup
- Encrypts sensitive DSN data using crypto module

#### Adapter Pattern (`internal/adapter`)

Pluggable interface for multi-database support:

```go
type DBAnalyzer interface {
    Connect(ctx context.Context, dsn string) error
    Close() error
    GetSlowQueries(ctx context.Context, opts QueryOpts) ([]SlowQuery, error)
    GetExplainPlan(ctx context.Context, query string, opts QueryOpts) (*ExplainPlan, error)
    GetIndexStats(ctx context.Context) ([]IndexStat, error)
    GetTableStats(ctx context.Context) ([]TableStat, error)
    GetDatabaseStats(ctx context.Context) (*DatabaseStats, error)
}
```

**PostgreSQL Adapter** (`internal/adapter/postgres.go`):

- Queries `pg_stat_statements` for slow query detection
- Runs read-only EXPLAIN ANALYZE for execution plans
- Collects `pg_stat_indexes` for index statistics
- Queries `information_schema` for table statistics

#### API Handler (`internal/api`)

Chi router with endpoint handlers:

- **Connection endpoints**: CRUD operations for database connections
- **Query endpoints**: List slow queries, stream live metrics via SSE
- **Query analysis endpoints**: EXPLAIN plans, historical analysis
- **Paste endpoint**: Parse and analyze MySQL slow log data

Middleware stack:

- Logger: Request/response logging via slog
- Recovery: Panic recovery with error response
- CORS: Allow cross-origin requests

Error responses use JSON format:

```json
{ "error": "connection failed: invalid DSN" }
```

#### Worker (`internal/worker`)

Background goroutine that:

1. Polls connected databases at `WORKER_INTERVAL_SECS` interval
2. Collects slow queries, index stats, table stats using adapters
3. Persists snapshots to Store
4. Signals API handlers for SSE real-time updates

Runs independently of HTTP request cycle.

#### Crypto Module (`internal/crypto`)

AES-256-GCM encryption/decryption:

- Encrypts DSN strings before storage (prevents plaintext secrets in DB)
- Decrypts DSN when establishing adapter connections
- Uses authenticated encryption (prevents tampering)

### 2. Frontend (React)

Built with Vite, TypeScript, shadcn/ui components.

**Directory Structure**:

- `src/components/` — Reusable UI components (shadcn/ui)
- `src/pages/` — Page layouts and views
- `src/hooks/` — Custom React hooks (useSSE, useQuery, etc.)
- `src/services/` — API client code
- `src/types/` — TypeScript type definitions

**Key Features**:

- Server-Sent Events (SSE) consumer for real-time metric streaming
- TanStack Table v8 for sortable/filterable query tables
- Recharts for real-time performance graphs
- Tailwind CSS for styling

### 3. Database (PostgreSQL)

Stores metadata and metrics snapshots:

- `connections`: Registered databases (id, name, dbType, encryptedDSN)
- `query_snapshots`: Periodic slow query snapshots
- `index_stats_snapshots`: Index usage statistics snapshots
- `table_stats_snapshots`: Table statistics snapshots (if implemented)

Migrations run automatically on startup.

## Data Flow

### Connection Creation

```bash
User Form (React)
    ↓ POST /api/connections
API Handler
    ↓
Encrypt DSN (Crypto module)
    ↓
Validate by connecting (Adapter)
    ↓
Store in DB (Store)
    ↓
Return connection ID (JSON)
```

### Real-time Query Metrics

```bash
Worker (Goroutine)
    ↓ Every 30s
For each connection:
  - Decrypt DSN
  - Query pg_stat_statements (Adapter)
  - Save snapshot (Store)
    ↓
SSE Endpoint (/api/connections/{id}/queries/stream)
    ↓
Push to React Dashboard (EventSource)
    ↓
Update charts and tables (React)
```

### Query Explain

```bash
User clicks "Explain" on query
    ↓ GET /api/connections/{id}/explain?query=...
API Handler
    ↓
Decrypt DSN (Crypto module)
    ↓
Run EXPLAIN ANALYZE (Adapter)
    ↓
Return plan (JSON)
    ↓
Display in modal (React)
```

## Security Considerations

- **DSN Encryption**: Sensitive database credentials encrypted with AES-256-GCM before storage
- **Parameterized Queries**: All DB queries use parameterized statements (pgx)
- **Read-only Adapter**: EXPLAIN transactions are read-only, no data modification
- **CORS**: Configured for browser-based access
- **No Auth Required** (MVP): Current implementation assumes trusted environment

## Performance Characteristics

| Operation             | Frequency                | Impact                              |
| --------------------- | ------------------------ | ----------------------------------- |
| Metrics polling       | Every 30s (configurable) | Low: read-only EXPLAIN queries      |
| SSE broadcasts        | Real-time per poll       | Low: depends on client count        |
| Store writes          | Every poll interval      | Low: indexed inserts to snapshots   |
| Encryption/decryption | Per adapter connect      | Negligible: one-time per poll cycle |

## Extensibility

**Add a new database adapter**:

1. Implement `DBAnalyzer` interface
2. Register in `adapter.NewAdapter()` switch
3. Add query patterns for target database's system catalog

**Add new metrics**:

1. Add query method to adapter interface
2. Create snapshot table in migrations
3. Update worker to collect new metric
4. Add API endpoint and React component

## Migration Strategy

Schema migrations in `migrations/` directory use embedded SQL files:

- `001_create_connections.sql`
- `002_create_query_snapshots.sql`
- `003_create_index_stats_snapshots.sql`

Migrations run once on server startup, tracked via `schema_version` table.
