# DBSight: Project Overview & Product Development Requirements

## Executive Summary

DBSight is a lightweight, single-binary database performance analyzer for PostgreSQL. It monitors slow queries via `pg_stat_statements`, visualizes execution plans, and provides index analysis—all with zero external dependencies beyond PostgreSQL. Built for DevOps engineers and database administrators who need real-time insights into query performance without complex infrastructure.

**Target Users:** Database administrators, platform engineers, DevOps teams managing PostgreSQL infrastructure.

**Key Value Proposition:** Fast, self-contained performance monitoring with zero learning curve—embed it in any PostgreSQL environment and start analyzing within minutes.

**License:** MIT

## Product Goals

1. **Minimize Setup Friction** — Single binary deployment, one environment variable for encryption, auto-running migrations
2. **Real-Time Observability** — Live query metric streaming via SSE, 30s polling interval (configurable)
3. **Actionable Insights** — EXPLAIN plans, index statistics, delta tracking to identify regression patterns
4. **Security by Default** — AES-256-GCM encryption for database credentials, no plaintext secrets in storage or API responses
5. **Extensibility** — Pluggable adapter interface; add MySQL, MariaDB, or other databases by implementing one interface

## Current Status: Production Ready

**Phases Completed (1–10):**

- [x] Project scaffold with Cobra CLI and dependency wiring
- [x] PostgreSQL schema, migrations, and store layer (pgxpool)
- [x] Database analyzer adapter interface + PostgreSQL implementation
- [x] Chi HTTP API with connection CRUD and middleware
- [x] Background worker with 30s polling scheduler
- [x] React frontend foundation (Vite, TypeScript, shadcn/ui)
- [x] Slow query dashboard with sortable tables and sparkline charts
- [x] EXPLAIN plan viewer — Direct mode + Paste JSON, collapsible tree with cost/scan warnings (Phase 08)
- [x] Index analysis — unused, duplicate, missing detection + SQL recommendations (Phase 09)
- [x] Docker multi-stage build, /healthz endpoint, docker-compose with healthchecks (Phase 10)

**Capabilities:**

- Create, test, edit, delete database connections
- View slow queries ranked by total time with execution delta tracking
- Real-time query metric streaming for live dashboard
- EXPLAIN plan visualization with sequential scan and row mismatch warnings
- Index analysis: identify unused/duplicate indexes, tables missing indexes, with generated DROP/CREATE SQL
- Parse MySQL slow log format (paste mode)
- Health check endpoint for container orchestration

## Functional Requirements

### F1: Connection Management

- Users can register PostgreSQL database connections with name, host, port, database, user, password
- Connections are tested before creation (latency reported to user)
- DSN stored encrypted in PostgreSQL, never transmitted in API responses
- Support multiple concurrent connections per dashboard instance
- Users can edit connection metadata and delete connections

### F2: Slow Query Detection & Ranking

- Collect `pg_stat_statements` metrics every 30 seconds (configurable)
- Rank queries by total execution time (calls × mean_time)
- Calculate execution delta per poll (identify recently slow queries vs. chronic issues)
- Support filtering by time range and minimum threshold
- Display query text, call count, execution time breakdown, and row count

### F3: Execution Plan Analysis

- Execute EXPLAIN ANALYZE (BUFFERS, FORMAT JSON) safely via read-only transactions
- Parse and display plan cost, rows affected, buffers used
- Support custom query execution with user-supplied SQL
- Show plan tree with node costs and timing
- Identify scan types and missing indexes from plan nodes

### F4: Index Analysis

- List all indexes per connection with usage statistics (`pg_stat_indexes`)
- Track index size, scan count, tuple reads/writes
- Identify unused indexes (idx_scan = 0)
- Identify missing indexes by analyzing execution plans for sequential scans

### F5: Real-Time Streaming

- Push query metric updates to connected clients via Server-Sent Events (SSE)
- Emit delta changes (new slow queries, timing changes) without full table reload
- Support multiple simultaneous clients per connection
- Graceful cleanup on client disconnect

### F6: Offline Analysis (Paste Mode)

- Accept MySQL slow log format via text input
- Parse query text, execution time, row counts
- Display analysis without live database connection
- Support JSON export for integration with external tools

### F7: Multi-Database Extensibility

- Define DBAnalyzer interface covering slow queries, EXPLAIN, index stats, table stats
- Implement PostgreSQL adapter to system catalog views
- Document interface for implementing MySQL, MariaDB, other adapters
- Allow adapter selection per connection without schema changes

## Non-Functional Requirements

### Performance

- **Latency:** API responses within 200ms (p95); SSE deltas within 100ms
- **Throughput:** Support up to 100 simultaneous metric streams
- **Polling Interval:** Default 30s, configurable 10s–300s (balance freshness vs. load)
- **Memory:** <100MB idle, <200MB under load (10 connections monitored)
- **Worker Concurrency:** Max 10 parallel adapter connections (semaphore-limited)

### Scalability

- **Connections:** Support 50+ registered connections per instance
- **Query History:** Retain 10k snapshots per connection (configurable retention)
- **Metrics Storage:** PostgreSQL backend, indexed for fast queries

### Availability

- **Graceful Degradation:** If target DB unreachable, worker continues polling other connections
- **Migration Durability:** On startup, run migrations idempotently; track version in schema_version table
- **Error Recovery:** Worker panics caught, logged, next poll proceeds normally

### Security

- **DSN Encryption:** AES-256-GCM with 32-byte key (64 hex chars), never stored plaintext
- **Input Validation:** Validate SQL queries, DSN format, numeric parameters
- **Read-Only Access:** EXPLAIN runs in read-only transactions; no data modification
- **SQL Injection Prevention:** Use parameterized queries (pgx $1, $2, etc.) exclusively
- **CORS:** Restrict to trusted origins (configurable)
- **No Authentication (MVP):** Assumes trusted environment; post-MVP add OAuth2/API tokens

### Reliability

- **Testing:** >80% code coverage for critical paths (connection, query collection, encryption)
- **Logging:** Structured logs via slog; Info, Warn, Error levels with context
- **Error Handling:** Explicit error wrapping and propagation; no silent failures
- **Graceful Shutdown:** Stop worker, close connections on SIGTERM/SIGINT

## Technical Constraints & Dependencies

### Tech Stack

| Layer         | Technology               | Rationale                                                             |
| ------------- | ------------------------ | --------------------------------------------------------------------- |
| Backend       | Go 1.26+                 | Compiled, minimal runtime; cross-platform; concurrency primitives     |
| API           | Chi router               | Lightweight, composable middleware, fast                              |
| Database      | PostgreSQL 14+           | Required for metadata + metrics storage; pg_stat_statements extension |
| Frontend      | React 19                 | Modern, Hooks-based, TypeScript support                               |
| UI Components | shadcn/ui + Tailwind CSS | Accessible, composable, no JS dependency bloat                        |
| Tables        | TanStack Table v8        | Headless, high performance, sortable/filterable                       |
| Encryption    | AES-256-GCM              | NIST-approved, authenticated encryption                               |
| Deployment    | Docker multi-stage       | Single image, minimal footprint                                       |

### Architectural Constraints

1. **Single Binary Requirement:** Frontend (React SPA) embedded in Go binary via `//go:embed web/dist`; no separate services
2. **Adapter Pattern:** All database access must go through `DBAnalyzer` interface; enables multi-DB support
3. **Worker Independence:** Metrics collector runs independently of request cycle; no blocking on slow target DBs
4. **Encryption Always:** All DSN storage must be encrypted; crypto key mandatory in config
5. **Migrations Embedded:** SQL migrations in Go via `embed.FS`; auto-run on startup

### External Dependencies

- **PostgreSQL Extension:** `pg_stat_statements` must be enabled on target databases (user responsibility)
- **Environment Variables:** PORT, DATABASE_URL, ENCRYPTION_KEY required; no config files
- **Go Modules:** Chi (router), pgx/v5 (database), Cobra (CLI), slog (logging)
- **Node/npm:** Only for frontend build; runtime uses embedded SPA

## Acceptance Criteria (MVP Phase)

### AC1: Connection Management

- [x] Create connection → validates DSN, stores encrypted
- [x] List connections → returns name, dbType, lastConnectedAt (no DSN)
- [x] Test connection → returns latency + pg_version
- [x] Update connection metadata
- [x] Delete connection → cascade deletes snapshots

### AC2: Query Detection

- [x] Worker collects pg_stat_statements every 30s
- [x] Snapshots persisted to PostgreSQL with timestamp
- [x] Queries ranked by (calls × mean_time)
- [x] Delta calculated as (current_total_time - previous_total_time)
- [x] No slow queries selected if pg_stat_statements unavailable

### AC3: Slow Query Dashboard

- [x] Table displays query text, calls, total time, delta, last execution
- [x] Sortable by any column
- [x] Filterable by time range, min duration
- [x] Sparkline charts show execution time trend
- [x] Click query → show detail drawer with full plan (if available)

### AC4: API Endpoints

- [x] GET /healthz (Phase 10)
- [x] GET /api/connections
- [x] POST /api/connections (create)
- [x] GET /api/connections/{id}
- [x] PUT /api/connections/{id} (update metadata)
- [x] DELETE /api/connections/{id}
- [x] POST /api/connections/{id}/test
- [x] GET /api/connections/{id}/queries
- [x] GET /api/connections/{id}/queries/stream (SSE)
- [x] POST /api/connections/{id}/explain (Phase 08)
- [x] GET /api/connections/{id}/indexes (Phase 09)
- [x] POST /api/paste/queries

### AC5: Security

- [x] ENCRYPTION_KEY required in config
- [x] DSN encrypted before storage
- [x] API responses exclude DSN and plain credentials
- [x] EXPLAIN runs in read-only transaction
- [x] Parameterized queries used throughout

### AC6: Testing

- [x] Encryption/decryption unit tests
- [x] Handler integration tests for CRUD
- [x] Worker scheduling tests
- [x] 70% code coverage

## Roadmap

### Phases 8–10: Completed 2026-02-22

- **Phase 08:** EXPLAIN plan viewer (direct + paste JSON mode, collapsible tree, scan warnings)
- **Phase 09:** Index analysis (unused, duplicate, missing detection; DROP/CREATE SQL generation)
- **Phase 10:** Docker 3-stage build, /healthz endpoint, docker-compose healthchecks, Makefile targets

### Phase 11: Authentication & RBAC (Post-MVP)

- OAuth2 / OIDC integration
- Role-based access (viewer, editor, admin)
- API token authentication
- Audit logging

### Phase 12: Advanced Metrics (Post-MVP)

- Table statistics (row count, dead rows, vacuum frequency)
- Lock wait analysis
- Long-running transaction tracking
- Replication lag monitoring

## Success Metrics

| Metric                  | Target     | Measurement                                                          |
| ----------------------- | ---------- | -------------------------------------------------------------------- |
| Time to Deploy          | <2 minutes | Copy binary, set env vars, run migrations                            |
| Query Detection Latency | <35s       | Time from slow query execution to dashboard visibility               |
| Dashboard Response Time | <200ms p95 | GET /api/connections/{id}/queries latency                            |
| User Retention          | >90%       | Users continue monitoring after first week                           |
| Feature Adoption        | >70%       | Users utilize EXPLAIN, index analysis, paste mode within first month |

## Risk Assessment

| Risk                                          | Likelihood | Impact                             | Mitigation                                             |
| --------------------------------------------- | ---------- | ---------------------------------- | ------------------------------------------------------ |
| PostgreSQL extension not enabled on target DB | Medium     | Worker polling fails silently      | Document requirement, add health check endpoint        |
| Slow worker blocking API requests             | Low        | UI unresponsive during collection  | Async worker, separate goroutine, timeouts             |
| Encryption key compromise                     | Low        | Stored credentials exposed         | Recommend key rotation, document key management        |
| Migration compatibility                       | Low        | Schema version mismatch on upgrade | Version tracking, idempotent migrations, rollback docs |

## Dependencies & Blockers

- **Backend Build:** Go 1.26+ compiler
- **Frontend Build:** Node.js 20+, npm
- **Database:** PostgreSQL 14+ accessible during migrations
- **Security:** 32-byte encryption key generation (openssl rand -hex 32)
- **Extensions:** pg_stat_statements enabled on target databases (user responsibility)

---

**Document Version:** 1.1
**Last Updated:** 2026-02-22
**Status:** Production Ready — Phases 1–10 Complete
