# Technical TODO — RegattaApi Review

> This file tracks **technical debt, security findings, and code quality issues** surfaced during codebase review.
> Feature-level TODOs (Zeitnahme, DRV imports, etc.) are in `Notes.md` — that file is the project roadmap.

---

## Security

- [ ] **[HIGH] Server-side capability enforcement missing**
  - The `Auth()` middleware verifies JWT but does NOT check `user_capability` per route.
  - Any authenticated user can bypass sidebar filtering and access `/internal/admin/*`, `/internal/zeitnahme/*`, etc.
  - Only `MetricsPage` (`internal/handlers/pages/pages.go:252`) manually checks capabilities.
  - **Fix**: Add a `RequireCap(cap string)` middleware that checks `c.Locals("capabilities")` and returns `Forbidden` if the cap is missing. Wire it into `wrapHandler` for protected routes, or add per-handler checks.
  - See: `internal/middleware/auth.go`, `internal/server/routing.go`

---

## Data Integrity

- [x] **Wrap `CreateMeldung` in a transaction**
  - `internal/crud/medlung.go` — inserts meldung + link_meldung_athlet rows atomically via `DB.WithTx` (reentrant).

- [x] **Wrap DRV import in a transaction**
  - `internal/handlers/api/v1/drv_import.go` — whole import now runs in a single `DB.WithTx` (atomic).

- [ ] **Make remaining crud queries tx-aware**
  - `DB.QueriesFromCtx(ctx)` silently falls back to the global pool when no tx is in context. Only the DRV import path and `CreateMeldung` are migrated; ~60 direct `DB.Queries.` call sites remain in `internal/crud/` (`pausen.go`, `users.go`, `users_group.go`, `verein.go`, `zeitnahme.go`, `obmann.go`, `buero.go`, and remaining spots in `athlet.go`/`rennen.go`/`medlung.go`).
  - Any of those called inside a future `DB.WithTx` block would run on a separate pooled connection outside the transaction, committing silently if the tx rolls back.
  - **Fix**: convert remaining `DB.Queries.` → `DB.QueriesFromCtx(ctx).` (mechanical) or add a lint/CI guard forbidding `DB.Queries.` in `internal/crud`.

---

## Code Quality

### Logging inconsistencies (fmt.Println → slog)

- [ ] **Replace `fmt.Println` with `slog` in `internal/service/leitung.go`**
  - Lines 45–46, 57, 78 use raw `fmt.Println`.
  - **Fix**: Replace with `slog.Info`/`slog.Warn`/`slog.Error`.

- [ ] **Replace `fmt.Println` with `slog` in `internal/handlers/api/v1/drv_import.go`**
  - Multiple `fmt.Println` calls throughout the file.
  - **Fix**: Replace with `slog` calls.

- [ ] **Replace `fmt.Println` with `slog` in `internal/utils/images.go`**
  - Any remaining `fmt.Println` calls should use `slog`.

### DRV import fragility

- [ ] **DRV import UUID comparison via `ClockSequence()` is fragile**
  - `internal/handlers/api/v1/drv_import.go:356–390` — uses `uuid.UUID.ClockSequence()` to compare recency of imported athletes. This is an unusual approach and may not reliably detect the most recent import.
  - **Consider**: Use a timestamp column or a dedicated import version tracking instead.

### `GetVerein` returns untyped fields

- [ ] **`GetVerein` returns `interface{}` for `GesKosten`/`GesZahlungen`**
  - `internal/sqlc/verein.sql.go` (queried by `internal/crud/verein.go:162`) returns a custom row type where `GesKosten`/`GesZahlungen` are typed as `interface{}` and must be type-asserted to `int64` (`verein.go:174–184`).
  - This is a sqlc/pgx limitation. Consider adding a dedicated SQL view or function that returns typed aggregates, or cast explicitly in SQL.

---

## Config & Documentation

- [ ] **Swagger docs are empty stubs**
  - `docs/swagger.json` (84 bytes) and `docs/swagger.yaml` (45 bytes) are nearly empty.
  - `just swagger-gen` is commented out of `just build` and `just build-air` (`Justfile:71,75`).
  - **Fix**: Either run `just swagger-gen` and add doc comments to API handlers, or remove `swagger-gen` from the Justfile and drop the docs directory.

- [ ] **Stale `.env` keys not read by config**
  - `.env` contains `ENV`, `APP_ENV`, and `PREFORK` which are never read by `internal/config/config.go`.
  - **Fix**: Either remove from `.env` or wire them into the config struct if they're needed.


