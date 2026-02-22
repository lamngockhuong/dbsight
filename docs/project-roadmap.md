# DBSight Project Roadmap

## Overview

This roadmap tracks the progression from MVP completion through production-ready features and enterprise capabilities. Each phase builds on the previous, maintaining backward compatibility where possible.

**Current Status:** Phases 1–11 Complete — **Multi-Database Support Ready**

## Phase 1-7: MVP (Completed)

**Status:** ✅ Complete

### Achievements

- Single-binary Go API + embedded React SPA
- Connection CRUD with AES-256-GCM encrypted DSN storage
- PostgreSQL adapter for slow query detection via pg_stat_statements
- 30-second polling worker with semaphore-limited concurrency
- Real-time metrics streaming via Server-Sent Events (SSE)
- Dashboard with sortable/filterable query tables
- TanStack Table v8 integration, Recharts visualization
- EXPLAIN query support via read-only transactions
- MySQL slow log paste mode for offline analysis
- Full TypeScript + React 19 frontend
- Docker multi-stage build
- 70% test coverage

### Key Metrics

- **MVP Duration:** 10 days (1 planner, 2 researchers, 3 fullstack devs, 1 tester, 1 code reviewer)
- **Lines of Code:** 3,250 (Go) + 1,400 (React)
- **Test Coverage:** >70%
- **Deployment:** Single 20MB binary

---

## Phase 8: EXPLAIN Plan Visualization (Complete)

**Status:** ✅ Complete | **Priority:** High | **Duration:** Completed 2026-02-22

### Delivered

- **Handler:** `internal/api/handlers/explain.go` — `RunExplain` (POST /api/connections/{id}/explain)
  - Accepts `{query, analyze_mode}` JSON body
  - Connects to target DB, runs `GetExplainPlan` via adapter in 30s timeout
- **Component:** `web/src/components/explain/explain-json-tree.tsx` — collapsible JSON tree with cost annotations, sequential scan warnings, row-estimate mismatch detection
- **Page:** `web/src/pages/explain-page.tsx` — Direct mode (run EXPLAIN via API) + Paste JSON mode, ANALYZE warning banner

### API Endpoint

```
POST /api/connections/{id}/explain
Body: { "query": "<SQL>", "analyze_mode": true }
```

---

## Phase 9: Index Analysis (Complete)

**Status:** ✅ Complete | **Priority:** High | **Duration:** Completed 2026-02-22

### Delivered

- **Models** (`internal/models/models.go`): `DuplicateIndex`, `Recommendation`, `IndexAnalysisResult`, `TableStat`
- **Adapter** (`internal/adapter/indexes.go`): `GetDuplicateIndexes()` method on PostgreSQL adapter
- **Handler** (`internal/api/handlers/indexes.go`): `GetIndexAnalysis` + `computeRecommendations`
  - Unused indexes: `idx_scan = 0`
  - Missing index candidates: tables with `seq_scans > 100` and `n_live_tup > 1000`
  - Duplicate indexes via `GetDuplicateIndexes()`
  - Generates DROP INDEX and CREATE INDEX SQL in recommendations
- **Page:** `web/src/pages/indexes-page.tsx` — summary cards, recommendations list, detail tables
- **Component:** `web/src/components/indexes/recommendations-list.tsx`

### API Endpoint

```
GET /api/connections/{id}/indexes
Returns: IndexAnalysisResult { unused_indexes, missing_candidates, duplicate_indexes, recommendations, captured_at }
```

---

## Phase 10: Docker + Deploy (Complete)

**Status:** ✅ Complete | **Priority:** High | **Duration:** Completed 2026-02-22

### Delivered

- **Store:** `Ping(ctx)` method added to `Store` interface and `PGStore` implementation
- **Router:** `/healthz` endpoint — checks DB connectivity via `Store.Ping()`, returns `{"status":"ok"}` or `503`
- **Dockerfile:** 3-stage build (Node → Go builder → Alpine runtime), non-root user, stripped binary
- **docker-compose.yml:** postgres healthcheck → migrate service → app with `HEALTHCHECK`
- **.dockerignore:** created to exclude dev artifacts
- **Makefile targets:** `docker-up`, `docker-down`, `generate-key`, `test`

### API Endpoint

```
GET /healthz
Returns: 200 {"status":"ok"} | 503 {"status":"error","error":"..."}
```

---

## Phase 10.5: Monorepo Restructure + Documentation Site (Complete)

**Status:** ✅ Complete | **Priority:** Medium | **Duration:** Completed 2026-02-22

### Delivered

- **Monorepo layout**: `web/` moved to `apps/web/`; root `pnpm-workspace.yaml` + `package.json` added
- **Docs site**: `apps/docs/` — Astro Starlight with EN + VI i18n, deployed to `dbsight.khuong.dev`
- **Makefile**: updated to use `pnpm --filter web`; added `dev-docs` and `build-docs` targets
- **Dockerfile**: updated for monorepo multi-stage build (`apps/web/dist`)
- **main.go embed**: updated to `//go:embed apps/web/dist`
- **GitHub Actions**: `deploy-docs.yml` — deploys `apps/docs` to GitHub Pages on push to `main`

---

## Phase 11: MySQL & MariaDB Support (Complete)

**Status:** ✅ Complete | **Priority:** High | **Duration:** Completed 2026-02-22

### Delivered

- **Adapters:** MySQL and MariaDB implementations of `DBAnalyzer` interface
  - `internal/adapter/mysql.go`, `mysql_slow_queries.go`, `mysql_explain.go`, `mysql_indexes.go`, `mysql_stats.go`
  - `internal/adapter/mariadb.go`, `mariadb_slow_queries.go`, `mariadb_explain.go`, `mariadb_indexes.go`, `mariadb_stats.go`
  - Shared utilities: `internal/adapter/mysqlcompat/helpers.go` (DSN builder, EXPLAIN JSON parsing)
- **Query Detection:**
  - MySQL/MariaDB: `performance_schema.events_statements_summary_by_digest` with picosecond-to-millisecond conversion
  - Supports MySQL 5.7+/8.0+ and MariaDB 10.x+ versions
- **EXPLAIN Plans:**
  - MySQL: `EXPLAIN FORMAT=JSON` and `EXPLAIN ANALYZE TREE FORMAT=JSON` (8.0.18+)
  - MariaDB: `EXPLAIN FORMAT=JSON` and `ANALYZE FORMAT=JSON`
  - Auto-detection of database version for correct syntax
- **Index Analysis:**
  - MySQL: queries `information_schema.statistics` with performance_schema metrics
  - MariaDB: JSON-based queries with `JSON_ARRAYAGG` for compatibility
- **Frontend:**
  - Connection form: database type selector (PostgreSQL/MySQL/MariaDB)
  - DSN builder with per-database format guidance
  - EXPLAIN JSON tree: automatic format detection for MySQL/MariaDB
  - New files: `dsn-builder.ts`, `mysql-explain-parser.ts`, `mariadb-explain-parser.ts`
- **Testing:**
  - Unit tests for EXPLAIN JSON parsing: `mysql_explain_test.go`, `mariadb_explain_test.go`
  - Integration tests with real MySQL/MariaDB instances

### Key Metrics

- **Adapters Added:** 2 (MySQL, MariaDB) with 5 files each
- **Shared Code:** 1 package with helpers for both adapters
- **Test Coverage:** EXPLAIN parsing verified for both formats
- **DSN Formats:** Documented with examples for all 3 DB types

---

## Phase 12: Authentication & RBAC (Post-MVP)

**Status:** 🔮 Future | **Priority:** Medium | **Estimated Duration:** 6 days

### Objective

Secure multi-user access with role-based permissions and audit logging.

### Features

- OAuth2 / OIDC integration (Google, GitHub, Azure AD)
- Role-based access control (Viewer, Editor, Admin)
- API token authentication for CI/CD integrations
- Audit logging (who accessed what, when)
- Session management with refresh tokens
- MFA support (TOTP)

### Key Considerations

- Maintain backward compatibility (allow unauthenticated mode for single-user deployments)
- Minimize performance impact on existing endpoints
- Support both centralized (OIDC) and local auth

---

## Phase 13: Advanced Metrics (Post-MVP)

**Status:** 🔮 Future | **Priority:** Low | **Estimated Duration:** 8 days

### Objective

Deeper database health monitoring beyond slow queries.

### Features

- **Table Statistics:** Row count, dead rows, vacuum/analyze frequency, bloat percentage
- **Lock Contention:** Identify long-running transactions, lock wait times, deadlocks
- **Replication Lag:** Monitor pg_stat_replication for standby lag
- **Connection Management:** Per-role connection usage, idle connections, connection pooler stats
- **Workload Analysis:** Read vs. write ratio, transaction isolation levels, abort rates

---

## Timeline & Resource Allocation

| Phase | Status      | Duration | Priority | Owner         | Start      | End        |
| ----- | ----------- | -------- | -------- | ------------- | ---------- | ---------- |
| 1-7   | ✅ Complete | 10 days  | —        | Team          | 2026-02-11 | 2026-02-21 |
| 8     | ✅ Complete | —        | High     | fullstack-dev | 2026-02-21 | 2026-02-22 |
| 9     | ✅ Complete | —        | High     | fullstack-dev | 2026-02-21 | 2026-02-22 |
| 10    | ✅ Complete | —        | High     | devops-eng    | 2026-02-21 | 2026-02-22 |
| 10.5  | ✅ Complete | —        | Medium   | devops-eng    | 2026-02-22 | 2026-02-22 |
| 11    | ✅ Complete | —        | High     | fullstack-dev | 2026-02-22 | 2026-02-22 |
| 12    | 🔮 Future   | 6 days   | Medium   | fullstack-dev | TBD        | TBD        |
| 13    | 🔮 Future   | 8 days   | Low      | fullstack-dev | TBD        | TBD        |

**Phases 1–11 completed 2026-02-22 — project now supports PostgreSQL, MySQL, and MariaDB with multi-database monitoring.**

## Success Metrics by Phase

### MVP (Complete)

- ✅ Single binary deployable
- ✅ Monitor 50+ connections concurrently
- ✅ Query detection latency <35s
- ✅ >70% test coverage
- ✅ Zero manual setup (migrations auto-run)

### Phase 8-10.5 (Production Ready — Complete)

- ✅ EXPLAIN plan viewer with collapsible JSON tree and scan warnings
- ✅ Index analysis: unused, duplicate, missing index detection with SQL recommendations
- ✅ Docker multi-stage build, /healthz endpoint, docker-compose with healthchecks
- ✅ Makefile targets: docker-up, docker-down, generate-key, test
- ✅ Monorepo (pnpm workspaces): `apps/web/` + `apps/docs/` (Starlight, EN+VI)
- ✅ GitHub Pages deployment via `deploy-docs.yml` → `dbsight.khuong.dev`

### Phase 11 (Multi-Database Support — Complete)

- ✅ MySQL 5.7+/8.0+ adapter with performance_schema slow query detection
- ✅ MariaDB 10.x+ adapter with performance_schema and JSON-based queries
- ✅ Per-database EXPLAIN FORMAT support (PostgreSQL JSON, MySQL/MariaDB JSON)
- ✅ Shared DSN builder and EXPLAIN parsers in `mysqlcompat/` package
- ✅ Frontend database type selector and DSN format guidance
- ✅ Documentation for all 3 DB types with setup prerequisites

### Phase 12-13 (Enterprise Ready)

- Support 100+ concurrent users
- RBAC with 3+ predefined roles
- Audit log retention >90 days
- Advanced metrics <500ms query latency
- SLA: 99.9% uptime

---

## Dependencies & Blockers

### Phase 8

- None (internal enhancement)

### Phase 9

- PostgreSQL 11+ (pg_stat_user_indexes)

### Phase 10

- Docker engine 20.10+
- Docker Compose v2+ (optional)
- Kubernetes 1.20+ (for K8s manifests)

### Phase 11

- OAuth2/OIDC provider (Google, GitHub, Azure, etc.)
- JWT library for token handling

### Phase 12

- PostgreSQL 12+ (enhanced replication stats)

---

## Breaking Changes & Migration Path

### MVP → Phase 8

- ExplainPlan model extended with RootNode field
- Migration: backward compatible (nil check)

### Phase 8 → Phase 9

- No schema changes
- New endpoints added (GET /api/indexes)

### Phase 9 → Phase 10

- Docker deployment; no API/database changes
- Existing deployments continue working

### Phase 10 → Phase 11

- Authentication middleware added (optional)
- API routes unchanged; new /auth/\* endpoints
- Unauthenticated mode supported for compatibility

---

## Feedback Loop & Community

- Monthly release notes (GitHub releases)
- User feedback survey (post-Phase 10)
- Feature request voting (GitHub discussions)
- Performance benchmarks (post-Phase 10)

---

**Document Version:** 1.3
**Last Updated:** 2026-02-22
**Next Review:** After Phase 12 planning
