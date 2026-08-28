# Technical TODO — RegattaApi Review

> This file tracks **technical debt, security findings, and code quality issues** surfaced during codebase review.
> Feature-level TODOs (Zeitnahme, DRV imports, etc.) are in `Notes.md` — that file is the project roadmap.

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

- [ ] Analyze and Fix different ways of handling things and enforce clear separation of abstraction layers

---

## Config & Documentation

- [ ] **Swagger docs are empty stubs**
  - `docs/swagger.json` (84 bytes) and `docs/swagger.yaml` (45 bytes) are nearly empty.
  - `just swagger-gen` is commented out of `just build` and `just build-air` (`Justfile:102,106`), and `swagger-fmt` out of `full-build` (`Justfile:110`).
  - **Fix**: either run `just swagger-gen` and add doc comments to API handlers, or remove `swagger-gen` from the Justfile and drop the docs directory.

- [ ] **Stale `.env` keys not read by config**
  - `ENV`, `APP_ENV`, and `PREFORK` are never read by `internal/config/config.go` (grep returns nothing). `PG_MAIL`/`PG_PW`/`GOOSE_*` are likewise only consumed by docker-compose/Justfile, not `config.go`.
  - **Fix**: either remove from `.env` or wire them into the config struct if needed.

---

## CI/CD

- [ ] Add a `just release` recipe that creates a git tag (and optionally triggers the docker build/push).
- [ ] On Release build docker image on GH Worker
  - Build and push docker image to GHCR.
  - No `.github/workflows` yet — only `release-docker`/`build-docker` Justfile recipes exist.
  - Have fixed semantic versioning workflow stored in a VERSION file

---

## Placeholder Pages (Work in progress)

Blank pages rendered with a `PageHeader` + "Work in progress ..." subtext. Implement the actual feature in each.

- [ ] **Zeitnahme → Vorsortierung** — `/internal/zeitnahme/vorsortierung` (`internal/templates/pages/zeitnahme/vorsortierung.templ`)
- [ ] **Zeitnahme → Wenderichter** — `/internal/zeitnahme/wenderichter` (`internal/templates/pages/zeitnahme/wenderichter.templ`)
- [ ] **Regattabüro → Kasse** — `/internal/regattabuero/kasse` (`internal/templates/pages/regattabuero/kasse.templ`)
- [ ] **Regattabüro → Startnummern Ausgabe/Rückgabe** — `/internal/regattabuero/startnummernausgabe` (`internal/templates/pages/regattabuero/startnummernausgabe.templ`)
- [ ] **Regattabüro → Änderungen von Obleuten** — `/internal/regattabuero/aenderungen_obleute` (`internal/templates/pages/regattabuero/aenderungen_obleute.templ`)

---
