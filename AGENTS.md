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

## Secrets Management (SOPS + age)

Two files, one plaintext and one encrypted:

- **`.env`** — local plaintext, **gitignored**. Loaded by `just` (`dotenv-load`) and `docker compose`.
- **`encrypt.env`** — committed, SOPS/age-encrypted copy. The single source of truth in the repo.
- **`.sops.yaml`** — `creation_rules` with `path_regex: \.env$` → age key `age1rlhtmq3vvhkwg8qd9g2zuz50wzxg8am9vzd56ksf3wkj8eulpctq5d0d5a`. The suffix regex matches both `.env` and `encrypt.env`.
- **flake.nix** — provides `encrypt-env` / `decrypt-env` shell commands (`writeShellApplication` wrappers around `sops`), plus `sops` and `age` in the dev shell.

```bash
just secrets-encrypt   # sops encrypt .env > encrypt.env
just secrets-decrypt   # sops decrypt encrypt.env > .env
```

Rules:

- **Never commit `.env`** — it is gitignored; only `encrypt.env` is tracked.
- After editing `.env`, run `just secrets-encrypt` and commit the updated `encrypt.env`.
- On a fresh clone, run `just secrets-decrypt` (needs the age private key at `~/.config/sops/age/keys.txt`).
- The `encrypt-env`/`decrypt-env` commands come from the flake; run `direnv reload` (or `nix develop`) after flake changes so they are on `PATH`.

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
- `public/` — served as static files (generated CSS, uploaded assets, WASM binaries)
- `internal/sqlc/` — generated DB query code from sqlc (do not edit; run `just sqlc-gen` to regenerate)
- `sqlc/` — source SQL files: `queries/*.sql` and `schemas/*.sql` (goose migrations)

## Important Gotchas

- `assets/` is **mounted** into `api-dev` — changes persist but air handles reload
- `internal/` and `main.go` **are** bind-mounted into `api-dev` — air rebuilds Go on any `.go` change; `.templ` changes trigger code regeneration
- `public/` contains only generated files (compiled CSS from Tailwind) — `just build` / `just tailwind-gen` produce them; no manual sync needed
- Inline `<script>` in templ files won't hot-reload — prefer `/assets/js/` files
- Toast `HX-Retarget: #toast-container` + default swap breaks toast visibility — must use `HX-Swap: beforeend`
- **Capability checks are UI-only** — see [TODO.md](./TODO.md). Enforce server-side in handler code or via a `RequireCap(cap string)` middleware.

## Tooling

- `templ` — generates Go from `.templ` files
- `sqlc` — generates DB layer from SQL
- `swag` — API docs (`just swagger-gen`)
- `air` — live reload (runs inside `api-dev` via `just watch`)
- `goose` — migrations via `just` commands
- `just tailwind-gen` — rebuild CSS (also run by air on CSS changes)
- `just wasm-build` — build Go WASM timekeeping client (`cmd/wasm/zeitnahme/` → `public/wasm/zeitnahme.wasm`)

## CRUD Layer Architecture

### Domain Types
Each CRUD entity in `internal/crud/` is a Go struct that **embeds** the sqlc-generated type and adds relational fields:

| File | Type | Embeds | Additional Fields |
|---|---|---|---|
| `athlet.go` | `Athlet` | `sqlc.Athlet` | `Rolle`, `Position`, `Verein`, `Meldungen`, `ErstesRennen` |
| `verein.go` | `Verein` | `sqlc.Verein` | `Athleten`, `GesKosten`, `GesZahlungen`, `Saldo` |
| `rennen.go` | `Rennen` | `sqlc.Rennen` | `Tag`, `NumMeldungen`, `NumAbteilungen`, `Meldungen` |
| `medlung.go` | `Meldung` | `sqlc.Meldung` | `Rennen`, `Verein`, `Athleten` |
| `users.go` | `User` | `sqlc.User` | `UserGroup` |
| `pausen.go` | `Pause` | `sqlc.Pause` | *(none — thin wrapper)* |

### `FromSqlc` Constructors
Every CRUD type has a constructor that wraps raw sqlc types into domain types:
```go
func AthletFromSqlc(a sqlc.Athlet) Athlet {
    return Athlet{Athlet: a}
}
```

### Lazy-Loading Getters
Getter methods check if a relational field is `nil` before returning; if nil, they emit a `slog.Warn` and fetch from DB:

- `(*Verein).GetAthleten()` — lazy-loads `v.Athleten` via `GetAllAthletenForVerein`
- `(*Rennen).GetMeldungen()` — lazy-loads `r.Meldungen` via `GetRennenMeldungen`

The `slog.Warn` log is intentional — it highlights call sites where post-fetching occurs so you can pre-load the data for better performance.

Stub getters (no DB fetch, return `nil, nil`):
- `(*Athlet).GetMeldungen()`
- `(*Meldung).GetAthleten()`
- `(*User).GetUserGroup()`

### PGType Helpers
Methods that safely unwrap nullable `pgtype` columns to Go native types (or `*string` / `*int`):
- `(*Athlet).GewichtStr()`, `(*Athlet).GeburtsdatumStr()`
- `(*Rennen).GetZusatz()`, `(*Rennen).GetKostenEur()`, `(*Rennen).GetRennabstand()`, `(*Rennen).GetStartzeit()`
- `(*Meldung).BemerkungStr()`, `(*Meldung).ZeitnahmeBemerkungStr()`, `(*Meldung).RechnungsNummerStr()`

### DB Query Patterns
Two patterns:

1. **Standalone functions** (most common) — every function starts with `getCtx(ctx)` for a 60s timeout, calls `DB.Queries.<QueryName>`, maps results:
```go
func GetAthlet(ctx context.Context, uuid uuid.UUID) (Athlet, error) {
    ctx, cancel := getCtx(ctx)
    defer cancel()
    // ...
}
```

2. **Method on domain type** — when the operation mutates the receiver:
```go
func (ath *Athlet) UpdateStartberechtigung(ctx context.Context, startberechtigt bool) error { ... }
```

The helper `isNoRowError` (`internal/crud/utils.go`) converts `pgx.ErrNoRows` to `apierr.ErrNotFound`.

### Critical: PostgreSQL NULL vs Go Types
pgx v5 **cannot** scan a SQL NULL into a non-nullable Go `string` or `int32`. JOINs producing NULL for NOT NULL columns cause scan errors. Prefer:
- Split queries (e.g., `GetRennenMinimal` + `GetRennenMeldungen`) over complex JOINs
- Ensure JOINs use nullable sqlc types for columns that may be NULL

## Error Handling & Logging

### No Silent Error Drops
**Never use `_ = err`.** Every error must be either:
- Returned to the caller
- Logged via `slog.Error` / `slog.Warn` / `slog.Debug`
- Passed to `handleAppError`

The `errcheck` linter in `golangci-lint` enforces this. Close/flush errors (gzip, file close) must be logged, not dropped.

### AppError System
All domain errors are `*handler.AppError` (wrapping `*apierr.AppError`) — see `internal/errors/app_error.go`:

| Constructor | Code | Status | Use Case |
|---|---|---|---|
| `handler.NotFound(msg)` | 404 | 404 | Resource not found |
| `handler.BadRequest(msg)` | 400 | 400 | Invalid input |
| `handler.Unauthorized(msg)` | 401 | 401 | Missing/invalid auth |
| `handler.Forbidden(msg)` | 403 | 403 | Insufficient permissions |
| `handler.NotAcceptable(msg)` | 406 | 406 | Not acceptable |
| `handler.ValidationError(fieldErrors)` | 1000 | 400 | Form validation errors |
| `handler.InternalError(msg)` | 500 | 500 | Unexpected errors |
| `handler.OK(msg)` | 200 | 200 | Success (triggers toast) |

`ValidationError` takes `map[string]string` for per-field error messages displayed inline next to form inputs.

### `handleAppError` Flow (`internal/server/error_handler.go`)
1. **Headers already written check** — if the compression middleware already committed the response, bail silently.
2. **Success (Code 200)** → `writeSuccessToast` — sets `HX-Retarget: #toast-container` + `HX-Swap: beforeend`, renders toast.
3. **API paths** (`/api/`) → `writeAPIError` — JSON `{code, status, message}`.
4. **Non-HTMX** → `writePageError` — full-page error via `ui_pages.Error()`.
5. **HTMX form errors** (POST/PUT/DELETE):
   - If `ae.FormComp != nil`: renders the form component inline (with field errors), then appends an OOB error toast.
   - Otherwise: sets `HX-Retarget: #toast-container` + `HX-Swap: beforeend`, returns an error toast.
6. **HTMX GET errors** → full-page error.

### Compression Middleware Caveat
The Compression middleware's `gw.Close()` (both deferred and explicit at end of handler) writes the gzip footer, which **commits HTTP 200** before any downstream error handler can react. If `HeadersWritten()` returns true in `handleAppError`, the error is silently dropped. **Fix errors upstream before they reach this point** — never rely on post-compression error handling.

## Handler & Server Patterns

### Handler Types & Registration
Routes are registered in `internal/server/routing.go` using four wrappers:

| Wrapper | File | Purpose | Middleware |
|---|---|---|---|
| `baseLayoutHandler(url, pageFunc)` | `renderer.go:17` | Full-page GET routes | Recovery + Compression + Logging + CORS + RateLimit + OptionalAuth + Timeout(60s) |
| `wrapHandler(handler, needAuth)` | `handler.go:111` | API + form POST/PUT/DELETE | Recovery + Compression + Logging + CORS + RateLimit + Timeout(30s) + [Auth if needAuth] |
| `templHandler(handler)` | `handler.go:139` | HTMX-only component endpoints | Returns 404 for non-HTMX requests |
| `wrapUIHandler(handler)` | `handler.go:168` | HTMX UI without auth | Same as templHandler but without HTMX-required check |

`wrapHandler` and `templHandler` use `handleAppError` from the middleware chain. `baseLayoutHandler` calls `handleAppError` directly after the chain.

### Custom Router & Path Parameters
The router (`internal/server/handler.go`) uses `{param}` wildcard syntax:
```go
r.Handle("GET", "/comp/zeitplan/{wettkampf}", templHandler(...))
```
Path parameters are stored in the request context using a custom key type `handler.CtxKeyPathParams` (`handler.CtxKey = string`) to prevent collisions:
```go
ctx := context.WithValue(req.Context(), handler.CtxKeyPathParams, params)
```
Extracted via `c.Param("key")` which falls back to `context.Value()` if not set on the Context directly.

### HTMX Conventions
- **Page navigation:** HTMX swaps `innerHTML` of `<main>` with `hx-push-url="true"` for history support — see `NavBar` for the pattern.
- **Success feedback:** Server sets `HX-Retarget: #toast-container` + `HX-Swap: beforeend` to **append** toasts. **Do not** use default innerHTML swap — that replaces the container and hides toasts.
- **Validation errors:** Return the form component with `fieldErrors` map via `ae.FormComp`, system renders it inline + appends an OOB error toast.
- **Redirects:** Set `HX-Redirect` header on success:
  ```go
  c.Writer.Header().Set("HX-Redirect", "/target")
  c.Writer.WriteHeader(http.StatusOK)
  return nil
  ```

## Form Components & Validation

### Reusable Input Components
Located in `internal/templates/components/input_components.templ`:

| Component | Description |
|---|---|
| `SubmitButton(text, color)` | Button with loading spinner |
| `DropDown(name, label, options, values, ...)` | Select with error state, placeholder, required |
| `LineInput(name, label, value, inpType, ...)` | Text/number/email/date input with error state |
| `FileUpload(name, label, ...)` | File input with error state |
| `RangeStepInput(name, label, value, ...)` | Range slider with step markers |
| `Toggle(name, label, checked, ...)` | Checkbox toggle |

All components accept a `fieldErrors map[string]string` parameter and display inline errors when the current field name matches an entry.

### Field-Level Validation Pattern
1. Define `fieldErrors := make(map[string]string)`.
2. Validate each field; on error, add to map.
3. If `len(fieldErrors) > 0`, return `handler.ValidationError(fieldErrors).WithForm(formComponent)`.
4. The error handler renders the form component inline (with error classes on invalid fields) and appends an OOB error toast.

### "Primary Form / Secondary Action" Pattern
For boolean toggles (e.g., `isActive`), the form uses a hidden primary input storing the inverted state and a visible secondary control that sets it:
- Hidden: `<input type="hidden" name="is_active" value="true">`
- Visible: Checkbox "user is not active" sets `is_not_active` → server reads `is_not_active`, stores `IsActive = !isNotActive`.

## Code Quality & Verification

- **Always run `just check` after significant changes** — this runs `go fmt ./...` → `golangci-lint run` → `go test ./... -v` → full build.
- **No `_ = err` drops** — every error must be handled. The `errcheck` linter catches violations.
- **Template changes** require regeneration: `just templ` (included in `just build`).
- **SQL query changes** require regeneration: `just sqlc-gen`.
- **Close/flush errors** (gzip, file, defer) must be logged via `slog.Warn`/`slog.Error`, never silently dropped.

## Import Aliases

| Alias | Module Path |
|---|---|
| `apierr` | `github.com/bata94/RegattaApi/internal/errors` |
| `DB` | `github.com/bata94/RegattaApi/internal/db` |
| `ui_pages` | `github.com/bata94/RegattaApi/internal/templates/pages` |
| `ui_components` | `github.com/bata94/RegattaApi/internal/templates/components` |
| `ui_layouts` | `github.com/bata94/RegattaApi/internal/templates/layout` |

## Key Conventions

- **UUID generation** — use `uuid.NewV7()` for time-ordered UUIDs (sqlc maps all `uuid` columns to `github.com/google/uuid.UUID`).
- **Context key type** — use `handler.CtxKey` string type for `context.WithValue` keys (not bare strings), to prevent collisions.
- **Config** — global singleton `config.C` loaded from env vars (`.env` + system env) in `main.go` before anything else. See `internal/config/config.go`.
- **Logging** — Go 1.21+ `log/slog` throughout. No third-party logger.
- **Server startup** — `main.go`: `config.Load()` → `DB.InitConnection()` → `utils.InitEmail()` → `http.ListenAndServe(addr, server.GetRouter())`.