# RegattaApi

Regatta management web application — athlete registration, race planning, timekeeping, DRV imports, and PDF generation for rowing regattas.

## Tech Stack

| Layer | Technology |
|---|---|
| Language | Go 1.26 |
| Database | PostgreSQL via pgx v5 |
| DB Layer | sqlc (codegen from SQL queries) |
| UI | Templ (Go-native templates) + HTMX |
| CSS | TailwindCSS 4 + daisyUI 5 |
| Auth | JWT (golang-jwt) + bcrypt |
| WebSocket | gorilla/websocket (live timekeeping) |
| WASM | Go js/wasm (zeitnahme timekeeping client) |
| Infrastructure | Docker multi-stage, Caddy reverse proxy |
| PDF | Gotenberg |
| Dev Environment | Air (live reload), Nix flakes, goose (migrations) |

## Architecture

```
RegattaApi/
├── main.go                  # Entrypoint
├── internal/
│   ├── config/              # Env-based app configuration
│   ├── db/                  # pgx pool, DB connection init
│   ├── sqlc/                # Generated: models, queries from sqlc (do not edit)
│   ├── crud/                # Domain types wrapping sqlc, lazy-loading getters, business queries
│   ├── errors/              # AppError struct, sentinel errors
│   ├── handler/             # Request context, middleware/error types
│   ├── handlers/            # HTTP handlers: pages/ (full-page), components/ (HTMX), api/v1/ (JSON)
│   ├── middleware/          # Recovery, Compression, Logging, CORS, RateLimit, Auth, Timeout
│   ├── server/              # Custom router, error handler, HTMX response helpers
│   ├── service/             # Business logic composing multiple CRUD calls
│   ├── templates/           # Templ components: layout/, components/, pages/, pdf/
│   └── utils/               # Email, files, PDF, metrics, string helpers
├── assets/                  # Frontend: JS, CSS source, icons, fonts, images
├── public/                  # Static output (compiled CSS, WASM binaries, uploaded files)
├── sqlc/                    # Source SQL: queries/ and schemas/ (goose migrations)
├── cmd/wasm/zeitnahme/      # Go WASM timekeeping client (builds to public/wasm/)
├── files/                   # User-uploaded files (DRV imports, PDFs)
├── docs/                    # Swagger docs (stub)
├── caddy/                   # Caddy data/certs volumes
├── bin/                     # Built binary (gitignored)
├── tmp/                     # Air build output (gitignored)
├── node_modules/            # npm dependencies (Tailwind, daisyUI)
├── Justfile                 # Task runner
├── sqlc.yaml                # sqlc config
├── .air.toml                # Air live-reload config
├── flake.nix                # Nix dev shell
├── docker-compose.yml       # Dev/prod service composition
└── Dockerfile               # Multi-stage build
```

## Quick Start

```bash
just dev           # starts api-dev + caddy at dev.localhost
just dev-debug     # starts api-dev on :8080 directly (no proxy)
just down          # stop all containers
```

File changes in `internal/` and `main.go` auto-reload via Air inside the container. Changes to `.templ` files trigger Templ regeneration, changes to `assets/css/input.css` trigger Tailwind rebuild.

## Key Just Commands

| Command | Action |
|---|---|
| `just check` | `fmt → lint → test → build` — run after significant changes |
| `just dev` | Start dev environment (api-dev + caddy) |
| `just build` | `templ generate` → `tailwindcss` → `wasm-build` → `go build` |
| `just full-build` | Full Docker-oriented build + sqlc-gen |
| `just sqlc-gen` | Regenerate sqlc Go code from SQL |
| `just db-up` | Run pending goose migrations |
| `just db-new NEW_MIG=<name>` | Create a new migration |
| `just db-status` | Show migration state |
| `just templ` | Regenerate Go from .templ files |
| `just tailwind-gen` | Rebuild CSS from Tailwind source |
| `just wasm-build` | Build Go WASM timekeeping client |
| `just lint` | `golangci-lint run` |
| `just test` | `go test ./... -v` |

## Key Architecture Decisions

- **HTMX-first** — minimal JS, server-driven UI. Page navigation swaps `<main>`, toasts use `HX-Retarget: #toast-container` + `HX-Swap: beforeend`.
- **No silent errors** — every error is returned, logged, or passed to `handleAppError`. The `errcheck` linter enforces this.
- **Lazy-loading getters** — CRUD domain types fetch relational data on first access (with `slog.Warn` to identify post-fetch penalties).
- **AppError system** — all errors flow through `handleAppError` which routes to toasts, JSON, or page errors depending on context.
- **Compression middleware caveat** — gzip footer commits HTTP 200 before error handlers can react; fix errors upstream.
- **pgx v5 NULL handling** — scanning SQL NULL into non-nullable Go types panics. Use split queries or nullable sqlc types for JOINs.

## Known Issues & Roadmap

See [TODO.md](./TODO.md) for technical debt, security findings, and outstanding work items.

Check errors during development: `docker compose logs -ftn 100 api-dev`
