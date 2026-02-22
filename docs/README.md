# DBSight Documentation

DBSight is a database performance analyzer web application that provides real-time insights into database query performance, index usage, and table statistics.

## Overview

**DBSight** helps development and DevOps teams identify slow queries, optimize database performance, and understand database patterns across PostgreSQL instances.

### Key Features

- Real-time query performance monitoring via pg_stat_statements
- Slow query detection and analysis with EXPLAIN plans
- Index usage statistics and recommendations
- Table statistics and bloat analysis
- Secure multi-database connection management (AES-256-GCM encrypted DSN storage)
- Server-Sent Events (SSE) for live metric streaming
- SPA frontend with responsive React UI

## Tech Stack

| Component         | Technology                                          |
| ----------------- | --------------------------------------------------- |
| **Backend**       | Go 1.26+, Chi router, pgx/v5, Cobra CLI             |
| **Frontend**      | React 18, Vite, TypeScript, shadcn/ui, Tailwind CSS |
| **Database**      | PostgreSQL (for metadata/metrics storage)           |
| **Security**      | AES-256-GCM encryption for DSN storage              |
| **Real-time**     | Server-Sent Events (SSE) for metrics streaming      |
| **Visualization** | Recharts, TanStack Table v8                         |

## Quick Start

### Prerequisites

- Go 1.26+
- Node.js 18+
- PostgreSQL 12+
- Docker & docker-compose (optional)

### Setup & Run

```bash
# 1. Start PostgreSQL
docker-compose up -d postgres

# 2. Copy environment template
cp .env.example .env
# Edit .env with your database credentials

# 3. Run database migrations
go run . migrate

# 4. Start backend server
go run . serve
# API available at http://localhost:42198

# 5. In another terminal, start frontend dev server
cd web
npm install
npm run dev
# Frontend available at http://localhost:5173
```

## Environment Variables

| Variable               | Default    | Description                            |
| ---------------------- | ---------- | -------------------------------------- |
| `PORT`                 | `42198`     | API server port                        |
| `DATABASE_URL`         | (required) | PostgreSQL connection string           |
| `ENCRYPTION_KEY`       | (required) | 32-byte hex for AES-256-GCM encryption |
| `WORKER_INTERVAL_SECS` | `30`       | Metrics polling interval (seconds)     |

### Example `.env`

```bash
PORT=42198
DATABASE_URL=postgres://dbanalyzer:secret@localhost:5432/dbanalyzer?sslmode=disable
ENCRYPTION_KEY=0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef
WORKER_INTERVAL_SECS=30
```

## Project Structure

```bash
.
├── main.go                    # Entry point, Cobra CLI, server wiring
├── internal/
│   ├── config/               # Environment configuration loading
│   ├── models/               # Domain types (SlowQuery, IndexStat, etc.)
│   ├── store/                # Data persistence layer & migrations
│   ├── adapter/              # Multi-DB adapter interface & implementations
│   ├── api/                  # Chi router, middleware, HTTP handlers
│   ├── worker/               # Background metrics collector goroutine
│   └── crypto/               # AES-256-GCM encryption/decryption
├── migrations/               # SQL migration files
├── web/                      # React SPA (Vite + TypeScript)
│   ├── src/
│   │   ├── components/       # UI components (shadcn/ui)
│   │   ├── pages/            # Page layouts and views
│   │   ├── hooks/            # Custom React hooks
│   │   ├── services/         # API client code
│   │   └── types/            # TypeScript type definitions
│   ├── index.html            # HTML entry point
│   ├── vite.config.ts        # Vite build configuration
│   └── package.json          # Frontend dependencies
├── docker-compose.yml        # Local PostgreSQL setup
└── docs/                     # Documentation
```

## CLI Commands

### `dbsight serve`

Starts the API server with embedded metrics worker.

```bash
go run . serve
```

Initializes PostgreSQL store, runs migrations, starts HTTP server on configured PORT, and launches background worker for metrics collection.

### `dbsight migrate`

Runs database migrations to initialize schema.

```bash
go run . migrate
```

## API Endpoints

Base URL: `http://localhost:42198/api`

| Method   | Endpoint                            | Description                   |
| -------- | ----------------------------------- | ----------------------------- |
| `GET`    | `/connections`                      | List all database connections |
| `POST`   | `/connections`                      | Create new connection         |
| `GET`    | `/connections/{id}`                 | Get connection details        |
| `PUT`    | `/connections/{id}`                 | Update connection             |
| `DELETE` | `/connections/{id}`                 | Delete connection             |
| `POST`   | `/connections/{id}/test`            | Test connection               |
| `GET`    | `/connections/{id}/queries`         | List slow queries             |
| `GET`    | `/connections/{id}/queries/stream`  | Stream live metrics (SSE)     |
| `GET`    | `/connections/{id}/queries/history` | Query history                 |
| `POST`   | `/paste/queries`                    | Parse slow log from paste     |

## Frontend Routes

| Path               | Component         | Purpose                           |
| ------------------ | ----------------- | --------------------------------- |
| `/`                | Home              | Dashboard overview                |
| `/connections`     | Connections List  | Manage database connections       |
| `/connections/:id` | Connection Detail | View connection metrics & queries |

## Further Reading

- [System Architecture](./system-architecture.md) — High-level design and component interactions
- [Code Standards](./code-standards.md) — Coding conventions and best practices
