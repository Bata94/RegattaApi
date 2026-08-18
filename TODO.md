# Technical TODO — RegattaApi Review

> This file tracks **technical debt, security findings, and code quality issues** surfaced during codebase review.
> Feature-level TODOs (Zeitnahme, DRV imports, etc.) are in `Notes.md` — that file is the project roadmap.

---

## Architecture / Web Framework

- [ ] **[HIGH] Extract a self-written, Fiber-like web framework on `net/http`**
  - Formalize the current bespoke stack into a reusable framework under `pkg/` (e.g. `pkg/webfw/`).
  - Today this is split across `internal/server/handler.go` (hand-rolled `{param}` matcher), `internal/handler` (`Context`, `AppError`), and four duplicated middleware stacks in `internal/server/handler.go` + `internal/server/renderer.go` (~152 routes).
  - API should feel like Fiber but stay `net/http`-compatible:
    - `app.Get("/path/:param", handler)` / `app.Post(...)` route registration, grouped/prefix routes.
    - `ctx.Next()` middleware chain + unified pipeline (Recovery / Compression / Logging / CORS / RateLimit / Timeout ± Auth / OptionalAuth) — replace the four hand-assembled wrapper stacks.
    - Context helpers: `ctx.JSON` / `ctx.HTML` / `ctx.Status`, path/query/body parsing, form/file helpers.
  - Preserve the existing HTMX toast/error-handling contract (`HX-Retarget: #toast-container` + `HX-Swap: beforeend`) and `httptest` compatibility.

---

## Security

- [ ] **[HIGH] Server-side capability enforcement missing**
  - The `Auth()` middleware verifies JWT but does NOT check `user_capability` per route.
  - Any authenticated user can bypass sidebar filtering and access `/internal/admin/*`, `/internal/zeitnahme/*`, etc.
  - Only `MetricsPage` (`internal/handlers/pages/pages.go:252-284`) manually checks capabilities.
  - **Fix**: Add a `RequireCap(cap string)` middleware that checks `c.Locals("capabilities")` and returns `Forbidden` if the cap is missing. Wire it into `wrapHandler`/`internalLayoutHandler` for protected routes, or add per-handler checks.
  - See: `internal/middleware/auth.go`, `internal/server/routing.go`

- [ ] **[HIGH] Unauthenticated WebSocket writes to timekeeping**
  - `/ws/zeitnahme` is registered as a raw `http.HandlerFunc` with **no Auth middleware** (`internal/server/routing.go:229`), and the upgrader uses `CheckOrigin: func() bool { return true }` (`internal/handlers/api/v1/zeitnahme_ws.go:19-23`).
  - `record_start` / `record_finish` / `assign_finish` messages write directly to the DB (`CreateZeitnahmeStart/Ziel`, `CreateZeitnahmeErgebnis`). Anyone on the network can inject race times.
  - **Fix**: require a valid token (or shared secret) on the WS handshake and restrict `CheckOrigin`.

- [ ] **[HIGH] Default JWT secret is a placeholder**
  - `internal/config/config.go:100` — `JWTSecret: getEnv("JWT_SECRET", "DO_NOT_USE_IN_PROD")`.
  - A production deploy that forgets to override it produces trivially forgeable tokens (admin caps included).
  - **Fix**: fail fast at startup when `JWT_SECRET` is missing or equals the placeholder.

- [ ] **[MEDIUM] CORS reflection with credentials**
  - `internal/middleware/cors.go:30-35,49-78` — default `CORS_ALLOWED_ORIGINS="*"`. `matchOrigin` treats `"*"` as "return the origin", so it sets `Access-Control-Allow-Origin: <attacker>` **plus** `Access-Control-Allow-Credentials: true`.
  - **Fix**: when `*` is configured, set `Allow-Origin: *` without credentials; never echo arbitrary origins with credentials.

- [ ] **[MEDIUM] Password-change endpoint can target arbitrary users**
  - `internal/handlers/components/components.go:346-396` — `ChangePasswordPost` only requires a valid token but acts on the `{uuid}` path param, not the caller's own `user_id`.
  - **Fix**: verify `c.Param("uuid") == caller user_id` (or require `allowed_admin`).

- [ ] **[MEDIUM] SMTP password logged in plaintext**
  - `internal/utils/email.go:47` — `slog.Error(... "options", emailOptions)` serializes the struct including `PW`.
  - **Fix**: redact `PW` (or stop logging `emailOptions`).

- [ ] **[LOW] `/api/v1/test` sends email to a hardcoded address**
  - `internal/handlers/api/v1/testing.go:8-24` — auth-only endpoint mails `bastian.sievers@gmail.com`. Leftover test/debug surface.
  - **Fix**: remove the route and handler.

- [ ] **[LOW] Metrics API not admin-gated**
  - `/metrics` page enforces `allowed_admin`, but `MetricsAPIHandler` (`/metricsApi`) is only `wrapHandler(..., true)`.
  - **Fix**: apply the same admin capability.

- [ ] **[LOW] Unauthenticated meldeergebnis HTML endpoint**
  - `GET /api/v1/leitung/meldeergebnis` is registered with `needAuth=false` (`internal/server/routing.go:181`) while sibling endpoints require auth.
  - **Fix**: make it `true`.

- [ ] **[LOW] API logout does not clear the auth cookie**
  - `internal/handlers/api/v1/auth.go:25-27` — `POST /api/auth/logout` returns `"Logout successful!"` but never deletes `auth_token`.
  - **Fix**: clear the cookie (or document that logout is UI-only).

- [ ] **[LOW] Spoofable client IP for rate limiting/logs**
  - `internal/handler/context.go:265-273` trusts `X-Forwarded-For` unconditionally; rate limiting keys on IP (`internal/middleware/ratelimit.go:81-93`).
  - **Fix**: trust the header only behind the proxy; add a per-user rate-limit key.

---

## UX and response System

- [ ] **Better System Error Handling — mask errors with a correlation UUID**
  - Generate a correlation **UUID** per error, `slog.Error` the full error **with that UUID**, and return only a generic message + UUID to the client so it can be traced in logs.
  - The `Error` type is a generic error with a `Code` and `Message` field.
    - [ ] Api Error Responses — `writeAPIError` should return `{code, status, message, errorId}`.
    - [ ] HTMX Component Responses — inline form errors + error toast should include the UUID, not the raw `err.Error()`.
    - [ ] WebPage Error Responses — full-page error should show generic message + UUID.
  - Current raw `err.Error()` leaks (overlaps this item): `internal/handlers/components/components.go:198,218,227,410,418,661,688,878`, `internal/handlers/api/v1/drv_import.go:49`, `internal/handlers/api/v1/setzung.go:142`, and several `api/v1/*.go` returning `handler.BadRequest(err.Error())`.

- [ ] On UI always return at least a Success or Error Toast on POST, PUT, DELETE, etc.

---

## Data Integrity

- [x] **Wrap `CreateMeldung` in a transaction**
  - `internal/crud/medlung.go` — inserts meldung + link_meldung_athlet rows atomically via `DB.WithTx` (reentrant).

- [x] **Wrap DRV import in a transaction**
  - `internal/handlers/api/v1/drv_import.go` — whole import now runs in a single `DB.WithTx` (atomic).

- [ ] **Make remaining crud queries tx-aware**
  - `DB.QueriesFromCtx(ctx)` silently falls back to the global pool when no tx is in context. Only the DRV import path and `CreateMeldung` are migrated; **~60** direct `DB.Queries.` call sites remain: `athlet.go` (7), `buero.go` (1), `medlung.go` (12), `obmann.go` (1), `pausen.go` (6), `rennen.go` (6), `users.go` (6), `users_group.go` (6), `verein.go` (3), `zeitnahme.go` (12).
  - Any of those called inside a future `DB.WithTx` block would run on a separate pooled connection outside the transaction, committing silently if the tx rolls back.
  - **Fix**: convert remaining `DB.Queries.` → `DB.QueriesFromCtx(ctx).` (mechanical) or add a lint/CI guard forbidding `DB.Queries.` in `internal/crud`.

- [ ] **[MEDIUM] Invoice number race + hardcoded year**
  - `internal/crud/verein.go:97-138` — `GetNextRechnungsnummer` reads all numbers then computes the next (non-atomic; concurrent calls collide) and hardcodes `"2024-"` as the prefix.
  - **Fix**: derive the year and use a DB sequence / advisory lock, or generate in the same transaction as `CreateRechnung`.

- [ ] **[MEDIUM] Invoice HTML path non-transactional; PDF path inconsistent**
  - `internal/handlers/api/v1/buero.go:114-174` (vs `:27-79`) — `KasseCreateRechnungHTML` stamps `SetMeldungRechnungsNummer` on every meldung then `CreateRechnung`; if the latter fails, meldungen are already stamped. `KasseCreateRechnungPDF` never creates a `rechnung` row or assigns numbers at all.
  - **Fix**: wrap in `DB.WithTx` and unify both paths.

- [ ] **[MEDIUM] `CreateZeitnahmeErgebnis` multi-step, non-atomic**
  - `internal/crud/zeitnahme.go:306-343` — insert ergebnis + `SetZeitnahmeStartVerarbeitet` + `SetZeitnahmeZielVerarbeitet` are three separate queries; partial failure leaves inconsistent "verarbeitet" flags.
  - **Fix**: wrap in a transaction.

- [ ] **[MEDIUM] Batch setters apply N individual UPDATEs without a transaction**
  - `internal/service/leitung.go:99-168` (`SetStartnummern`, `ResetStartnummern`, `SetZeitplan`), `internal/handlers/api/v1/setzung.go:88-131`, `internal/crud/medlung.go:326-343`.
  - Partial application on mid-loop error (e.g. Setzung partially assigned).
  - **Fix**: `DB.WithTx` around each batch.

- [ ] **[LOW] `defer conn.Release()` before nil check**
  - `internal/db/db.go:136-139` — if `pool.Acquire` fails, `conn` is nil and the deferred `conn.Release()` panics.
  - **Fix**: check `err` before deferring.

- [ ] **[LOW] Substring panic on short rechnungsnummer**
  - `internal/crud/verein.go:108` — `rNrStr := r[l-3:l]` panics if any stored number is < 3 chars.
  - **Fix**: length-guard.

---

## Code Quality

### Logging inconsistencies (fmt.Println → slog)

- [x] **Replace `fmt.Println` with `slog` in `internal/service/leitung.go`**
- [x] **Replace `fmt.Println` with `slog` in `internal/handlers/api/v1/drv_import.go`**
- [x] **Replace `fmt.Println` with `slog` in `internal/utils/images.go`**
- [x] **Replace `fmt.Printf` with `slog` in `internal/utils/files.go`**
- [x] **Replace `fmt.Println` with `slog` in `internal/templates/components/image.templ`**

### DRV import fragility

- [ ] **DRV import UUID comparison via `ClockSequence()` is fragile**
  - `internal/handlers/api/v1/drv_import.go:356-390` — uses `uuid.UUID.ClockSequence()` to compare recency of imported athletes. Unusual and may not reliably detect the most recent import.
  - **Consider**: a timestamp column or dedicated import version tracking.

### `GetVerein` returns untyped fields

- [ ] **`GetVerein` returns `interface{}` for `GesKosten`/`GesZahlungen`**
  - `internal/sqlc/verein.sql.go` (queried by `internal/crud/verein.go:162`) returns a custom row type where `GesKosten`/`GesZahlungen` are typed `interface{}` and must be type-asserted to `int64` (`verein.go:174-184`).
  - sqlc/pgx limitation. Consider a dedicated SQL view/function returning typed aggregates, or cast explicitly in SQL.

### Misc

- [ ] **[MEDIUM] `string(rune(int(x)+'0'))` misrenders multi-digit numbers**
  - `internal/handlers/api/v1/buero.go:152,154` — converts `StartNummer`/`Kosten` to a single character via rune arithmetic (12 → form-feed, 10 → newline). Start numbers go up to 350.
  - **Fix**: `strconv.Itoa`.

- [ ] **[LOW] Timekeeping day mismatch**
  - `internal/handlers/api/v1/zeitnahme.go:91` (`GenerateEndZeit` hardcodes `crud.TagSa`) vs `zeitnahme_ws.go:291` (`assign_finish` uses `config.C.Zeitnahme.GetCurrentTag()`).
  - **Fix**: use the configured day consistently.

- [ ] **[LOW] Dead code**
  - `internal/errors/app_error.go:76-81` (`ErrWrongRefreshToken`, `ErrTokenInvalid`, `ErrTimeParse` unused); `internal/utils/fmt_struct.go` (`FormatStruct`, `FormatListOfStructs` unused). Refresh-token machinery has no refresh flow.
  - **Fix**: remove or implement.

- [ ] **[LOW] Duplicated helpers and middleware stacks**
  - `internal/handler/context.go:176-184` (`JSON` vs `JSONOk` identical); `internal/server/handler.go:119-197` + `renderer.go:19-128` (four wrapper funcs each re-declare the same middleware list).
  - **Fix**: extract a single stack builder and drop `JSONOk`. (Also addressed by the framework extraction item.)

- [ ] **[LOW] Magic numbers/strings**
  - `internal/handlers/components/components.go:515` (`MAX_STARTNUMMER=350`), `internal/crud/verein.go:136` (`"2024-"`), `internal/handlers/api/v1/buero.go:69` (`"MRG Regatta 24"`), `internal/handlers/api/v1/drv_import.go:497-507` (100/310/321 thresholds).
  - **Fix**: extract named constants; derive dates from `time.Now()`.

- [ ] **[LOW] `context.Background()`/`context.TODO()` bypass cancellation**
  - `internal/handlers/api/v1/zeitnahme_ws.go:78,177,221,260`, `internal/crud/verein.go:55,217,235`, `internal/crud/rennen.go:128` — lazy-loaders and WS handlers use background contexts, losing request timeouts.
  - **Fix**: thread the request ctx through.

---

## Config & Documentation

- [ ] **Swagger docs are empty stubs**
  - `docs/swagger.json` (84 bytes) and `docs/swagger.yaml` (45 bytes) are nearly empty.
  - `just swagger-gen` is commented out of `just build` and `just build-air` (`Justfile:102,106`), and `swagger-fmt` out of `full-build` (`Justfile:110`).
  - **Fix**: either run `just swagger-gen` and add doc comments to API handlers, or remove `swagger-gen` from the Justfile and drop the docs directory.

- [ ] **Stale `.env` keys not read by config**
  - `ENV`, `APP_ENV`, and `PREFORK` are never read by `internal/config/config.go` (grep returns nothing). `PG_MAIL`/`PG_PW`/`GOOSE_*` are likewise only consumed by docker-compose/Justfile, not `config.go`.
  - **Fix**: either remove from `.env` or wire them into the config struct if needed.

- [ ] **[LOW] No `golangci-lint` config**
  - No `.golangci.yml`; `just lint` runs with defaults despite AGENTS.md describing `errcheck` enforcement (which isn't actually pinned).
  - **Fix**: add a `.golangci.yml` enabling `errcheck` (and the other described linters).

---

## CI/CD

- [ ] Add a `just release` recipe that creates a git tag (and optionally triggers the docker build/push).
- [ ] On Release build docker image on GH Worker
  - Build and push docker image to GHCR.
  - No `.github/workflows` yet — only `release-docker`/`build-docker` Justfile recipes exist.

- [ ] **[LOW] `open-firewall`/`close-firewall` sudo recipes**
  - `Justfile:181-189` open port 8080 via `iptables`; accidental use exposes the app.
  - **Fix**: remove or gate behind a warning.

---

## Placeholder Pages (Work in progress)

Blank pages rendered with a `PageHeader` + "Work in progress ..." subtext. Implement the actual feature in each.

- [ ] **Zeitnahme → Vorsortierung** — `/internal/zeitnahme/vorsortierung` (`internal/templates/pages/zeitnahme/vorsortierung.templ`)
- [ ] **Zeitnahme → Wenderichter** — `/internal/zeitnahme/wenderichter` (`internal/templates/pages/zeitnahme/wenderichter.templ`)
- [ ] **Regattabüro → Kasse** — `/internal/regattabuero/kasse` (`internal/templates/pages/regattabuero/kasse.templ`)
- [ ] **Regattabüro → Startnummern Ausgabe/Rückgabe** — `/internal/regattabuero/startnummernausgabe` (`internal/templates/pages/regattabuero/startnummernausgabe.templ`)
- [ ] **Regattabüro → Änderungen von Obleuten** — `/internal/regattabuero/aenderungen_obleute` (`internal/templates/pages/regattabuero/aenderungen_obleute.templ`)
- [X] **Regattaleitung → E-Mail senden** — `/internal/regattaleitung/email` (`internal/templates/pages/regattaleitung/email.templ`)
- [X] **Regattaleitung → Startnummernbereiche** — `/internal/regattaleitung/startnummern/bereich` (`internal/templates/pages/regattaleitung/startnummern.templ`, `StartnummernBereich`)
- [x] **Regattaleitung → Vereine verwalten** — `/internal/regattaleitung/vereine` (`internal/templates/pages/regattaleitung/vereinverwaltung.templ`)

---
