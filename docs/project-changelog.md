# DBSight Project Changelog

All notable changes to DBSight are documented here.

---

## [Unreleased]

### Added — UI Enhancements (Responsive, Theme, Notifications)

- **Responsive mobile layout**: Sheet-based drawer sidebar for mobile devices
- **Dark/Light/System theme support**: Custom ThemeProvider with localStorage persistence (replaces next-themes)
- **Theme toggle**: DropdownMenu component to switch between Light, Dark, and System themes
- **Toast notifications**: Sonner integration for CRUD feedback (success, error, loading states)
- **Breadcrumb navigation**: Breadcrumbs component for page hierarchy visualization
- **Sidebar enhancements**:
  - New `sidebar-nav.tsx` for reusable navigation links
  - New `sidebar-footer.tsx` with author info, GitHub link, docs link, and app version
  - New `mobile-sidebar.tsx` for responsive drawer implementation
- **New UI components**: Sheet, DropdownMenu, Breadcrumb (via shadcn/ui)
- **Lucide icons**: Added for theme toggle and UI enhancements
- **Removed dependency**: next-themes (replaced with custom solution)

---

## [1.1.0] — 2026-02-22

### Added — Monorepo Restructure + Documentation Site

- **Monorepo layout**: Restructured from flat layout to pnpm workspaces monorepo
  - `web/` moved to `apps/web/` (package name: `web`)
  - New `apps/docs/` — Astro Starlight documentation site with EN + VI i18n
  - Root `pnpm-workspace.yaml` and `package.json` added
- **main.go embed directive**: Updated from `//go:embed web/dist` to `//go:embed apps/web/dist`
- **Makefile targets**: Updated all `pnpm` calls to `pnpm --filter web`; added `dev-docs` and `build-docs`
- **Dockerfile**: Updated multi-stage build for monorepo layout (`apps/web/dist`)
- **GitHub Actions**: Added `.github/workflows/deploy-docs.yml` — deploys `apps/docs` to GitHub Pages on push to `main` (path filter: `apps/docs/**`)
- **Docs site URL**: `dbsight.khuong.dev`

### Changed

- Development frontend command changed from `cd web && pnpm run dev` to `pnpm --filter web dev`
- Build command changed from `cd web && pnpm run build` to `pnpm --filter web build`

---

## [1.0.0] — 2026-02-22

### Added — Production Ready (Phases 1–10)

- Single-binary Go API server with embedded React SPA
- Connection CRUD with AES-256-GCM encrypted DSN storage
- PostgreSQL adapter for slow query detection via `pg_stat_statements`
- 30-second polling background worker with semaphore-limited concurrency (max 10)
- Real-time metrics streaming via Server-Sent Events (SSE)
- Dashboard with sortable/filterable slow query tables (TanStack Table v8)
- Recharts visualization for query execution trends
- EXPLAIN plan viewer: collapsible JSON tree, cost annotations, sequential scan warnings, row-estimate mismatch detection
- Index analysis: unused indexes, duplicate index detection, missing index candidates; generates DROP/CREATE INDEX SQL recommendations
- MySQL slow log paste mode for offline analysis
- Docker multi-stage build (Node → Go builder → Alpine runtime, non-root user)
- `docker-compose.yml` with postgres healthcheck → migrate → app
- `/healthz` endpoint with `Store.Ping()` check
- Makefile targets: `build`, `docker-build`, `docker-up`, `docker-down`, `generate-key`, `test`
- Full TypeScript + React 19 frontend with shadcn/ui + Tailwind CSS v4
- AES-256-GCM encryption for stored DSNs (`ENCRYPTION_KEY` env var)
- Idempotent SQL migrations tracked via `schema_version` table
- 70% test coverage on critical paths

---

**Document Version:** 1.0
**Last Updated:** 2026-02-22
