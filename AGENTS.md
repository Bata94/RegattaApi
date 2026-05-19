# RegattaApi — Agent Guide

## Dev Environment (Docker-Centric)

- `just dev` — starts `api-dev` + `caddy` containers
- `just dev-debug` — starts only `api-dev` (direct access, no proxy)
- `just down` — stop all containers

**File changes auto-reload** — no manual rebuild needed:
- `.go` / `.templ` / `.tpl` → air rebuilds Go + regenerates templ code inside `api-dev`
- `assets/css/input.css` → air runs `just tailwind-gen` on save
- `assets/js/*.js` → air triggers SSE, browser reloads the file

**Check for errors:** `docker compose logs -ftn 100 api-dev`

## Build Commands (for reference)

`just build` runs `templ → tailwind-gen → go build` in sequence. Full Docker build: `just full-build`.

## Database Migrations

```bash
just db-up           # run all pending
just db-new NEW_MIG=<name>  # create migration
just db-status       # show migration state
```

Migrations use goose with postgres. Driver/dbstring come from `.env`.

## HTMX-First Architecture

- **Prefer HTMX over JS.** The goal is minimal JS.
- **Page navigation:** use HTMX to swap `innerHTML` of `<main>` (or equivalent root element), with `hx-push-url="true"` to update browser history — see NavBar for the pattern.
- **Dynamic updates:** use `hx-get`, `hx-post`, etc. on elements rather than adding event listeners in JS.
- **Toast responses:** server sets `HX-Retarget: #toast-container` + `HX-Swap: beforeend` to **append** toasts. Frontend `toast.js` auto-dismisses after 5s (excludes `alert-error`). **Do not** use default innerHTML swap — that replaces the container and hides toasts.

## JS Implementation Priority

When JS is needed, prefer in this order:
1. **File in `/assets/js/`** — loaded via `<script src>` in `base_ui_layout.templ`. Hot-reloaded by air.
2. **`templ Scripts()` function** — generates a `<script>` tag referencing a static asset. Still external, but embedded in the templ code generation.
3. **Inline `<script>` tags** — only as a last resort. Not hot-reloaded by air; requires container restart to take effect.

## Key Directories

- `internal/handler/` — HTTP handlers
- `internal/handlers/` — page-level handlers
- `internal/server/` — router setup, middleware, htmx response helpers
- `internal/templates/` — templ components/layouts/pages
- `assets/js/` — external JS (hot-reloaded by air)
- `assets/css/` — tailwind source (`input.css` → `public/css_global.css`)
- `public/` — served as static files (generated CSS, uploaded assets)
- `sqlc/` — generated DB query code (run `just sqlc-gen` to regenerate)

## Important Gotchas

- `assets/` is **mounted** into `api-dev` — changes persist but air handles reload
- `public/` is **not** mounted — needs manual rebuild or `just full-build`
- Inline `<script>` in templ files won't hot-reload — prefer `/assets/js/` files
- Toast `HX-Retarget: #toast-container` + default swap breaks toast visibility — must use `HX-Swap: beforeend`

## Tooling

- `templ` — generates Go from `.templ` files
- `sqlc` — generates DB layer from SQL
- `swag` — API docs (`just swagger-gen`)
- `air` — live reload (runs inside `api-dev` via `just watch`)
- `goose` — migrations via `just` commands
- `just tailwind-gen` — rebuild CSS (also run by air on CSS changes)