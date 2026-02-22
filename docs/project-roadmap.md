# DBSight Project Roadmap

## Overview

This roadmap tracks the progression from MVP completion through production-ready features and enterprise capabilities. Each phase builds on the previous, maintaining backward compatibility where possible.

**Current Status:** MVP Complete (Phases 1–7) — **Ready for Phase 8**

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

## Phase 8: EXPLAIN Plan Visualization (Planned)

**Status:** 📋 Planned | **Priority:** High | **Estimated Duration:** 5 days

### Objective

Provide visual query plan tree with cost breakdown and performance analysis to help developers optimize slow queries.

### User Stories

#### US1: Query Plan Tree Rendering

- Display EXPLAIN plan as interactive tree diagram
- Show node types (Seq Scan, Index Scan, Hash Join, Sort, etc.)
- Display costs per node (estimated vs. actual)
- Show rows processed per node
- Highlight expensive nodes (estimated cost > 10% of total)

#### US2: Node Details & Metrics

- Click node → side panel with full details
- Buffer usage breakdown (hits vs. reads)
- Planning time vs. execution time comparison
- Identify n-loops (nested loop joins)
- Show execution timing per node

#### US3: Query Optimization Recommendations

- Detect sequential scans → suggest indexes
- Identify inefficient joins (nested loops on large datasets)
- Highlight suboptimal sort operations
- Recommend statistics updates if plan is based on old stats
- Flag expensive function calls in WHERE clause

#### US4: Plan Comparison Tool

- Load two plans side-by-side (before/after optimization)
- Highlight cost differences
- Show improvement percentage
- Track metrics over time (estimated vs. actual accuracy)

### Technical Implementation

**Frontend:**

- Create `web/src/components/explain/plan-tree.tsx` — React tree visualization
- Create `web/src/components/explain/node-detail-panel.tsx` — Node details
- Create `web/src/components/explain/recommendation-panel.tsx` — Optimization tips
- Integrate Recharts for cost distribution pie chart
- Add comparison mode toggle in explain-page.tsx

**Backend:**

- Extend ExplainPlan model with node hierarchy (parent/child relationships)
- Add cost analysis logic in adapter/explain.go
- Implement recommendation engine (scan type detection, join analysis)
- API endpoint: GET /api/connections/{id}/explain?query=<SQL>&compare=<prev-id> (optional)

**Data Model:**

```go
type ExplainNode struct {
    NodeType       string          // "Seq Scan", "Index Scan", etc.
    EstimatedCost  float64
    ActualCost     float64
    EstimatedRows  int64
    ActualRows     int64
    BuffersHit     int64
    BuffersRead    int64
    ExecutionTime  float64
    Children       []*ExplainNode  // Nested nodes
    Details        map[string]any  // Node-specific info
}

type ExplainPlan struct {
    // ... existing fields ...
    RootNode       *ExplainNode
    Recommendations []*Recommendation
}

type Recommendation struct {
    Type       string  // "missing_index", "seq_scan", "inefficient_join"
    Severity   string  // "critical", "high", "medium", "low"
    Message    string
    Suggestion string
}
```

### Success Criteria

- [x] Plan tree renders correctly for single-query explains
- [x] Node details accessible via click/hover
- [x] Recommendations displayed for >80% of query anti-patterns
- [x] Plan comparison works for 90% of use cases
- [x] <300ms rendering time for plans with <50 nodes

### Risks & Mitigations

| Risk                                        | Mitigation                                      |
| ------------------------------------------- | ----------------------------------------------- |
| Complex nested plans hard to visualize      | Collapsible tree view, zoom controls            |
| Recommendation engine gives false positives | Manual review, confidence scores, feedback loop |
| Performance slow for large plans            | Lazy load nodes, virtualization, memoization    |

---

## Phase 9: Index Analysis Dashboard (Planned)

**Status:** 📋 Planned | **Priority:** High | **Estimated Duration:** 4 days

### Objective

Help database administrators identify unused indexes, track index bloat, and suggest missing indexes based on query patterns.

### User Stories

#### US1: Unused Index Detection

- List all indexes per connection with usage stats (pg_stat_user_indexes)
- Highlight indexes where idx_scan = 0 (never scanned)
- Filter by age, size, table
- Show candidates for removal with safety checks (no foreign keys, etc.)
- One-click generate DROP INDEX script

#### US2: Index Bloat Analysis

- Calculate index size vs. table size ratio
- Identify bloated indexes (high ratio = candidates for REINDEX)
- Show last vacuum/analyze timestamps
- Recommend VACUUM ANALYZE or REINDEX based on bloat percentage

#### US3: Missing Index Suggestions

- Analyze EXPLAIN plans from recent slow queries
- Detect Seq Scans on large tables with filter conditions
- Suggest CREATE INDEX statements
- Estimate potential performance improvement (based on cost reduction)
- Track which suggestions were implemented

#### US4: Index Performance Heatmap

- Matrix: tables × index performance (size, scan count, hit ratio)
- Color-coded: green (efficient), yellow (bloated), red (unused)
- Click cell → detail view with recommendations
- Time-series trend (index growth, scan count trend)

### Technical Implementation

**Frontend:**

- Create `web/src/pages/indexes-page.tsx` — Main index analysis page (may already exist as stub)
- Create `web/src/components/indexes/unused-indexes-table.tsx` — Sortable table
- Create `web/src/components/indexes/bloat-analysis.tsx` — Bloat heatmap
- Create `web/src/components/indexes/missing-indexes.tsx` — Suggestions list
- Create `web/src/components/indexes/index-detail-modal.tsx` — Full details + actions

**Backend:**

- Extend DBAnalyzer interface: GetIndexStats, GetMissingIndexCandidates
- Implement PostgreSQL logic in adapter/indexes.go
  - Query pg_stat_user_indexes for usage stats
  - Calculate index size from pg_indexes, pg_relation_size()
  - Cross-reference against EXPLAIN plans to suggest indexes
- API endpoints:
  - GET /api/connections/{id}/indexes — List with stats
  - GET /api/connections/{id}/indexes/missing — Missing index suggestions
  - GET /api/connections/{id}/indexes/{indexName}/drop-script — DROP INDEX statement
  - GET /api/connections/{id}/indexes/{indexName}/reindex-script — REINDEX statement

**Data Model:**

```go
type IndexStat struct {
    IndexName      string    // e.g., "idx_users_email"
    TableName      string
    IndexSize      int64     // bytes
    TableSize      int64     // bytes
    IdxScan        int64     // number of times scanned
    IdxTuplRead    int64     // tuples read from index
    IdxTuplFetch   int64     // tuples fetched from index
    CreatedAt      time.Time
    LastIndexScan  *time.Time
    BloatPercent   float64   // (index_size - estimated_true_size) / index_size * 100
}

type MissingIndexSuggestion struct {
    TableName      string
    Columns        []string
    Where          string  // optional WHERE clause
    EstimatedGain  float64 // % improvement in slow queries
    Frequency      int     // how many slow queries would benefit
    CreateStatement string  // full CREATE INDEX statement
}
```

### Success Criteria

- [x] Detect 100% of unused indexes (idx_scan = 0)
- [x] Bloat analysis accurate for 90% of cases
- [x] Missing index suggestions for queries using Seq Scan with filters
- [x] Generate safe DROP INDEX scripts (no manual CASCADE needed)
- [x] Dashboard loads in <1s for connections with 100+ indexes

### Risks & Mitigations

| Risk                           | Mitigation                                                    |
| ------------------------------ | ------------------------------------------------------------- |
| False positive missing indexes | Manual review, highlight confidence, test before implement    |
| REINDEX blocking writes        | Document risk, recommend off-peak hours, show table lock time |
| Index stats outdated           | Trigger ANALYZE refresh button, document pg_stat_reset        |

---

## Phase 10: Docker Production Deployment (Planned)

**Status:** 📋 Planned | **Priority:** High | **Estimated Duration:** 3 days

### Objective

Provide production-ready deployment patterns with security, health checks, and observability.

### User Stories

#### US1: Multi-Stage Docker Build

- Optimize image size (Go binary ~10MB, React dist ~200KB, total <50MB)
- Build frontend in Node stage, Go binary in Go stage
- Runtime image: distroless or alpine (minimal attack surface)
- Support arm64 and amd64 architectures

#### US2: Health Checks & Graceful Shutdown

- Implement /health endpoint (check DB connectivity, worker status)
- SIGTERM handler: drain SSE connections, stop worker, close DB
- Readiness probe: DB migrations complete
- Liveness probe: worker heartbeat <1 min old

#### US3: Production Configuration Guide

- Reverse proxy setup (nginx, Caddy) with SSL/TLS
- Environment variable best practices (secret management with 1Password, HashiCorp Vault, etc.)
- Log aggregation (stderr to syslog, JSON structured logging)
- Database backup strategy (pg_dump frequency, retention)
- Monitoring setup (Prometheus metrics, Grafana dashboards)

#### US4: Docker Compose for Local Development

- postgres service (14+, pg_stat_statements enabled)
- dbsight service (linked to postgres)
- pgAdmin optional (for manual DB inspection)
- Volume mounting for frontend hot reload in dev

#### US5: Kubernetes Manifests (Optional)

- Deployment spec with resource limits
- Service definition (ClusterIP, NodePort options)
- ConfigMap for non-secret env vars
- Secret object for ENCRYPTION_KEY
- StatefulSet option for persistent metrics storage

### Technical Implementation

**Docker:**

- Update Dockerfile:

  ```dockerfile
  # Stage 1: Build React
  FROM node:20-alpine AS frontend
  WORKDIR /app/web
  COPY web/ .
  RUN npm ci && npm run build

  # Stage 2: Build Go
  FROM golang:1.26-alpine AS builder
  WORKDIR /app
  COPY . .
  COPY --from=frontend /app/web/dist ./web/dist
  RUN go build -o bin/dbsight .

  # Stage 3: Runtime
  FROM gcr.io/distroless/base-debian12
  COPY --from=builder /app/bin/dbsight /
  EXPOSE 42198
  ENTRYPOINT ["/dbsight", "serve"]
  ```

**Health Checks:**

- Add handler: GET /health → returns {status: "ok", db: "connected", worker: "active"}
- Docker HEALTHCHECK: `curl -f http://localhost:42198/health || exit 1`
- Update main.go: graceful shutdown handler on SIGTERM

**Observability:**

- Add Prometheus metrics endpoint (optional): GET /metrics
- Export: worker_polls_total, worker_errors_total, api_requests_duration_seconds, etc.
- Structured logging: include request_id, connection_id, operation in all logs

**Documentation:**

- Create docs/deployment-guide.md (this PR)
  - Docker run command with env vars
  - docker-compose.yml for dev
  - nginx reverse proxy config (SSL termination)
  - Database backup recommendations
  - Monitoring setup (Prometheus scrape config)
  - Kubernetes manifests (YAML templates)

### Success Criteria

- [x] Docker image builds successfully, runs without errors
- [x] Image size <100MB
- [x] Health check responds within 100ms
- [x] Graceful shutdown drains connections within 30s
- [x] docker-compose up works with zero manual DB setup
- [x] Deployment guide covers 80% of production scenarios

### Risks & Mitigations

| Risk                               | Mitigation                                                     |
| ---------------------------------- | -------------------------------------------------------------- |
| Secret management complexity       | Document multiple approaches (env vars, Vault, Docker secrets) |
| Database connection pooling issues | Test with load generator, tune pgxpool maxConns                |
| Cold start delays                  | Cache adapter connections per target DB                        |

---

## Phase 11: Authentication & RBAC (Post-MVP)

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

## Phase 12: Advanced Metrics (Post-MVP)

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
| 8     | 📋 Planned  | 5 days   | High     | fullstack-dev | 2026-02-21 | 2026-02-26 |
| 9     | 📋 Planned  | 4 days   | High     | fullstack-dev | 2026-02-26 | 2026-03-02 |
| 10    | 📋 Planned  | 3 days   | High     | devops-eng    | 2026-03-02 | 2026-03-05 |
| 11    | 🔮 Future   | 6 days   | Medium   | fullstack-dev | TBD        | TBD        |
| 12    | 🔮 Future   | 8 days   | Low      | fullstack-dev | TBD        | TBD        |

**Total:** ~26 days post-MVP to reach production-ready state (Phases 8–10)

## Success Metrics by Phase

### MVP (Complete)

- ✅ Single binary deployable
- ✅ Monitor 50+ connections concurrently
- ✅ Query detection latency <35s
- ✅ >70% test coverage
- ✅ Zero manual setup (migrations auto-run)

### Phase 8-10 (Production Ready)

- Dashboard renders plans <300ms
- Index analysis accurate 90%+ of the time
- Docker image <100MB
- Health checks respond within 100ms
- Graceful shutdown within 30s
- Documentation covers 80% of deployment scenarios

### Phase 11-12 (Enterprise Ready)

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

**Document Version:** 1.0
**Last Updated:** 2026-02-21
**Next Review:** After Phase 8 completion
