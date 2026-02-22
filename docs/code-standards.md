# Code Standards Overview

DBSight maintains consistent coding standards across Go backend and TypeScript frontend to ensure maintainability, security, and developer productivity.

## Universal Principles

- **YAGNI**: Don't add code unless immediately needed
- **KISS**: Keep solutions simple and readable
- **DRY**: Avoid code duplication; extract shared logic
- **Security First**: Parameterized queries, input validation, secret encryption
- **Error Handling**: Explicit propagation and structured logging
- **Testing**: >80% coverage for critical paths
- **Documentation**: Comment exported APIs, keep examples current

## Go Backend

### Standards

- **Logging**: `log/slog` structured logs (Info, Warn, Error levels)
- **Database**: `pgx/v5` parameterized queries (always use `$1, $2` placeholders)
- **Interfaces**: Define in consuming package; enables dependency injection
- **Context**: First parameter; enables cancellation, timeouts, tracing
- **Error Handling**: Wrap errors with `fmt.Errorf("op: %w", err)`
- **Security**: AES-256-GCM encrypt DSN before storage; read-only EXPLAIN txns
- **File Structure**: Packages under `internal/` by domain (config, models, store, adapter, api, worker, crypto)
- **Naming**: Packages lowercase; functions PascalCase (exported), camelCase (unexported); JSON fields snake_case

### Key Files

- `main.go` — CLI entry point, dependency wiring
- `internal/config/` — Environment configuration
- `internal/models/` — Domain types
- `internal/store/` — Persistence interface & PostgreSQL impl
- `internal/adapter/` — Multi-DB interface & implementations
- `internal/api/` — Chi router, handlers, middleware
- `internal/worker/` — Background metrics collector
- `internal/crypto/` — Encryption utilities

## TypeScript Frontend

### Standards

- **Strict Mode**: Enabled; catches type errors at compile time
- **Components**: Functional with hooks; keep <200 LOC per component
- **Files**: kebab-case names (e.g., `query-detail-drawer.tsx`)
- **Components**: PascalCase (e.g., `<QueryDetailDrawer />`)
- **Hooks**: Prefix `use` (e.g., `useQueryHistory()`)
- **Types**: PascalCase interfaces
- **Styling**: Tailwind CSS + shadcn/ui components
- **API**: Centralize calls in `src/services/api.ts`
- **State**: Use `useEffect` for data fetching; handle loading/error/cleanup
- **Imports**: Use `@/` path aliases (configured in tsconfig)

### Key Directories

- `src/components/ui/` — shadcn/ui base components
- `src/components/layout/` — Layout wrappers
- `src/components/queries/`, `connections/` — Feature components
- `src/hooks/` — Custom React hooks
- `src/services/` — API client code
- `src/types/` — TypeScript definitions
- `src/pages/` — Route-level components

## API Contract

### Requests

All endpoints accept JSON; use `?` query params for filtering.

### Success Response

```json
{
  "id": 123,
  "name": "Production DB",
  "db_type": "postgres",
  "created_at": "2026-02-21T12:00:00Z",
  "updated_at": "2026-02-21T12:00:00Z"
}
```

### Error Response

```json
{"error": "connection failed: invalid DSN"}
```

### Status Codes

- `200 OK` — Success
- `201 Created` — Resource created
- `400 Bad Request` — Invalid input
- `404 Not Found` — Resource not found
- `500 Internal Server Error` — Server error

## Testing

- Go: Place `*_test.go` in same package; use table-driven tests
- TypeScript: React Testing Library; focus on user interactions, not implementation
- Mock APIs via MSW (Mock Service Worker)

## Further Reading

- [System Architecture](./system-architecture.md) — Component interactions, data flow
- [README](./README.md) — Quick start, project overview
