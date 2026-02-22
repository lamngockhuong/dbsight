# DBSight Frontend

React SPA for DBSight — a database performance analyzer targeting PostgreSQL.

## Tech Stack

- **React 19** + **TypeScript** + **Vite 7**
- **Tailwind CSS v4** + **shadcn/ui** (Radix UI primitives)
- **TanStack Table v8** — query data tables
- **Recharts** — sparkline charts
- **React Router v7** — client-side routing
- **Biome** — linting & formatting (replaces ESLint + Prettier)

## Getting Started

```bash
pnpm install
pnpm run dev        # Vite dev server at http://localhost:5173
```

Vite proxies `/api` requests to `http://localhost:8080` (Go backend).

## Scripts

| Command             | Description                    |
| ------------------- | ------------------------------ |
| `pnpm run dev`      | Start Vite dev server with HMR |
| `pnpm run build`    | Type-check + production build  |
| `pnpm run preview`  | Preview production build       |
| `pnpm run lint`     | Run Biome check                |
| `pnpm run lint:fix` | Run Biome check with auto-fix  |
| `pnpm run format`   | Format with Biome              |

## Project Structure

```text
src/
├── api/            — Typed fetch wrapper for API endpoints
├── components/
│   ├── connections/ — Connection form & list
│   ├── layout/      — Sidebar + Layout shell
│   ├── queries/     — Query table, detail drawer, sparklines
│   └── ui/          — shadcn/ui components (do not edit manually)
├── hooks/           — Custom hooks (connections, queries, SSE)
├── pages/           — Route pages
├── types/           — TypeScript interfaces mirroring Go models
├── lib/             — Utility functions
├── App.tsx          — Router wiring
└── main.tsx         — Entry point
```

## Notes

- `@/` path alias configured in `tsconfig.json` + `vite.config.ts`
- shadcn/ui components: add via `npx shadcn@latest add <component>`
- Production build output (`dist/`) is embedded into the Go binary via `//go:embed`
