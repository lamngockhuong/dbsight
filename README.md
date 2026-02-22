# DBSight

Database performance analyzer for PostgreSQL, MySQL, and MariaDB. Monitor slow queries, visualize EXPLAIN plans, and track index usage — all from a single binary.

![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go&logoColor=white)
![React](https://img.shields.io/badge/React-19-61DAFB?logo=react&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-14+-4169E1?logo=postgresql&logoColor=white)
![MySQL](https://img.shields.io/badge/MySQL-5.7+-005A87?logo=mysql&logoColor=white)
![MariaDB](https://img.shields.io/badge/MariaDB-10.x+-C0765F?logo=mariadb&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-green)

**Documentation:** [dbsight.khuong.dev](https://dbsight.khuong.dev)

## Features

- **Slow Query Detection** — polls `pg_stat_statements` (PostgreSQL), `performance_schema` (MySQL/MariaDB) every 30s, ranks by total execution time with delta tracking
- **Live Dashboard** — real-time updates via Server-Sent Events (SSE), no page refresh needed
- **EXPLAIN Plans** — run EXPLAIN safely with per-database format support (PostgreSQL JSON, MySQL FORMAT=JSON, MariaDB ANALYZE FORMAT=JSON)
- **Index Analysis** — identify unused indexes and missing index opportunities
- **Paste Mode** — analyze slow query logs offline without a live database connection
- **Multi-Database** — monitor multiple PostgreSQL, MySQL, and MariaDB instances from one dashboard
- **Secure** — DSN credentials encrypted with AES-256-GCM, never exposed via API

## Quick Start

### Prerequisites

- Go 1.26+
- Node.js 20+
- One or more supported databases:
  - PostgreSQL 14+ (with `pg_stat_statements` extension enabled)
  - MySQL 5.7+ or 8.0+ (with `performance_schema` enabled)
  - MariaDB 10.x+ (with `performance_schema` enabled)
- Docker & Docker Compose (optional)

### Using Docker Compose

```bash
docker-compose up -d postgres
```

### Setup

```bash
# Generate a 32-byte encryption key
export ENCRYPTION_KEY=$(openssl rand -hex 32)

# Configure database
export DATABASE_URL="postgres://dbsight:secret@localhost:5499/dbsight?sslmode=disable"

# Run migrations
go run . migrate

# Start the server (API + worker)
go run . serve
```

### Frontend Development

```bash
pnpm install          # Install all workspace dependencies
pnpm --filter web dev # Vite dev server on :5173, proxies /api to :42198
```

### Documentation Site

```bash
pnpm --filter docs dev   # Starlight dev server on :4321
pnpm --filter docs build # Build static docs site
```

### Production Build

```bash
make build     # Builds frontend, then Go binary → bin/dbsight
./bin/dbsight serve
```

Or with Docker:

```bash
make docker-build
docker run -e DATABASE_URL=... -e ENCRYPTION_KEY=... -p 42198:42198 dbsight:latest
```

## Architecture

This is a **pnpm workspaces monorepo**:

```
dbsight/
├── apps/web/        # React SPA (Vite + shadcn/ui)
├── apps/docs/       # Starlight documentation site (EN + VI)
├── internal/        # Go backend packages
├── migrations/      # SQL migration files
├── main.go          # Entry point — embeds apps/web/dist into binary
└── docker-compose.yml
```

```
┌─────────────────────────────────────────────┐
│              Go Binary (dbsight)            │
│                                             │
│  ┌──────────┐  ┌────────┐  ┌────────────┐   │
│  │ Chi API  │  │ Worker │  │ Embedded   │   │
│  │ Server   │  │ (30s)  │  │ React SPA  │   │
│  └────┬─────┘  └───┬────┘  └────────────┘   │
│       │            │                        │
│  ┌────┴────────────┴──────┐                 │
│  │    Store (pgxpool)     │                 │
│  └────────────┬───────────┘                 │
└───────────────┼─────────────────────────────┘
                │
        ┌───────┴───────┐
        │  PostgreSQL   │ (app metadata + metrics)
        └───────────────┘
                │
        ┌───────┴───────┐
        │  Target DBs   │ (via DBAnalyzer adapter)
        └───────────────┘
```

The single binary serves the API, background worker, and React SPA. The worker collects metrics from target databases via the adapter interface — extensible to MySQL and others.

## Tech Stack

| Layer     | Technology                                             |
| --------- | ------------------------------------------------------ |
| Backend   | Go 1.26+, Chi router, pgx/v5, Cobra CLI                |
| Frontend  | React 19, Vite, TypeScript, shadcn/ui, Tailwind CSS v4 |
| Data      | TanStack Table v8, Recharts                            |
| Database  | PostgreSQL (metadata storage)                          |
| Security  | AES-256-GCM encrypted DSN storage                      |
| Real-time | Server-Sent Events (SSE)                               |
| Deploy    | Docker multi-stage build                               |
| Docs      | Astro Starlight, i18n (EN + VI), Pagefind search       |

## Environment Variables

| Variable               | Default | Description                             |
| ---------------------- | ------- | --------------------------------------- |
| `PORT`                 | `42198` | HTTP server port                        |
| `DATABASE_URL`         | —       | PostgreSQL connection string for app DB |
| `ENCRYPTION_KEY`       | —       | 64 hex chars (32 bytes) for AES-256-GCM |
| `WORKER_INTERVAL_SECS` | `30`    | Metrics polling interval in seconds     |

## Target Database Setup

### PostgreSQL

Enable `pg_stat_statements` on the databases you want to monitor:

```sql
-- postgresql.conf
shared_preload_libraries = 'pg_stat_statements'

-- Then run:
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;
```

### MySQL 5.7+ / 8.0+

Ensure `performance_schema` is enabled (usually on by default):

```sql
-- Check if it is enabled
SHOW GLOBAL VARIABLES LIKE 'performance_schema';
-- Should return 'ON'

-- If disabled, add to my.cnf and restart:
[mysqld]
performance_schema = ON
```

### MariaDB 10.x+

Enable `performance_schema` in the configuration:

```sql
-- Check if it is enabled
SHOW GLOBAL VARIABLES LIKE 'performance_schema';
-- Should return 'ON'

-- If disabled, add to my.cnf and restart:
[mysqld]
performance_schema = ON
```

## API

| Method         | Endpoint                                | Description                       |
| -------------- | --------------------------------------- | --------------------------------- |
| GET/POST       | `/api/connections`                      | List / create connections         |
| GET/PUT/DELETE | `/api/connections/{id}`                 | Get / update / delete connection  |
| POST           | `/api/connections/{id}/test`            | Test connection (returns latency) |
| GET            | `/api/connections/{id}/queries`         | Latest slow queries with deltas   |
| GET            | `/api/connections/{id}/queries/stream`  | SSE live query updates            |
| GET            | `/api/connections/{id}/queries/history` | Historical snapshots              |
| POST           | `/api/paste/queries`                    | Parse slow query log text         |

## Project Status

- [x] Project scaffold + config
- [x] Database schema + store layer
- [x] DB adapter interface + PostgreSQL implementation
- [x] API server + connection management
- [x] Background worker + query endpoints
- [x] React frontend foundation
- [x] Slow query dashboard UI
- [x] EXPLAIN plan visualization (custom tree renderer)
- [x] Index analysis dashboard
- [x] Docker production deployment
- [x] Monorepo restructure (pnpm workspaces)
- [x] Documentation site (Starlight, EN + VI)

## License

MIT
