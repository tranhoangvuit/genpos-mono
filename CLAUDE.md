# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository overview

GenPOS is a multi-tenant retail Point of Sale system. The repo holds three deployable apps plus shared infra, all wired around a single set of proto definitions and a PostgreSQL/PowerSync backend.

- `backend/` — Go ConnectRPC server (multi-tenant SaaS API)
- `frontend/` — TanStack Start (React) web app, served at :3032
- `genpos-desk/` — Tauri v2 desktop POS (Rust + React, offline-first SQLite, syncs via PowerSync)
- `powersync/` — PowerSync service config + sync rules (used by both clients)
- `postgres/` — first-boot init script
- `deploy/` — production Compose stack (Caddy, prebuilt images, Cloudflare TLS)
- `docs/` — `deploy-ubuntu.md` is the production deployment walkthrough

## Top-level commands (root `Makefile`)

| Target | What it does |
|---|---|
| `make infra` | brings up Postgres 17 (:3033, `wal_level=logical`), Redis 7, PowerSync (:3034), and the one-shot `migrate` container |
| `make infra-down` / `make infra-logs` | stop / tail logs |
| `make backend` | runs Go server on :3031 with `POWERSYNC_ENDPOINT=http://localhost:3034` |
| `make proto` | regenerates Go code from `backend/proto/` (`buf generate` in `backend/`) |
| `make frontend` / `make frontend-install` | dev server (:3032) / `pnpm install` |
| `make desk` / `make desk-install` / `make desk-proto` | Tauri dev / pnpm install / regenerate TS bindings for the desktop app |
| `make clean` | `docker compose down -v` + nuke frontend `node_modules`/`.output` |

`docker compose up -d` (what `make infra` runs) brings up two databases: `genpos_dev` (default) and `genpos_test` (for integration tests). The `migrate` service runs Atlas migrations on every `up` and is a dependency of both `backend` and `powersync`, so PowerSync never starts against an unmigrated DB.

## Backend (`backend/`)

Go module: `github.com/genpick/genpos-mono/backend` (Go 1.25). Core stack: ConnectRPC, pgx/v5, sqlc, Atlas migrations, `goforj/wire` for DI, `samber/oops` for errors, `go.uber.org/mock` for mocks.

### Layered architecture

Strict dependency direction: `handler → usecase → domain.gateway` (interfaces) ← `infrastructure/datastore` (implementations).

- `cmd/server/main.go` — entrypoint; loads config via `kelseyhightower/envconfig`, calls `app.InitializeApp`, starts HTTP/2 cleartext on `:3031` (configurable via `PORT`).
- `internal/app/` — wire graph. `wire.go` (build tag `wireinject`) is the source of truth for the dependency graph; `wire_gen.go` is the hand-written equivalent kept until wire supports Go 1.25 — **edit both** when adding providers.
- `internal/handler/grpc/` — one Connect handler per service (auth, catalog, customer, order, …). Each is registered in `app.NewHTTPHandler` with three interceptors applied: DB (transaction injection), Auth (JWT), Permission (RBAC, see `interceptor/authz.go`).
- `internal/usecase/` — business logic; depends only on `domain/gateway` interfaces.
- `internal/domain/` — entities + gateway interfaces (the ports).
- `internal/infrastructure/datastore/` — pgx-backed implementations of gateways. `tenant_db.go` resolves the per-tenant connection used inside interceptors. SQL queries live in `query/` and are compiled to typed Go via `sqlc generate` (output in `infrastructure/datastore/sqlc/`).
- `pkg/auth/` — argon2 password hashing, JWT issuance, PowerSync JWKS signing, refresh-token rotation. `pkg/database/postgres.go` wraps `pgxpool`.

### Two database connections, two roles

`App` holds **both `DB` and `AuthDB`** (`internal/app/app.go`). Migration `20260425000001_separate_db_roles.sql` introduces three roles seeded by `database/bootstrap/`: `app_runner` (general queries), `app_auth` (auth tables), `powersync_user` (logical replication). The auth pool uses the more privileged role; treat them as separate connection pools when wiring new gateways.

### Proto + codegen

Proto sources live in `backend/proto/genpos/v1/*.proto` and are the single source of truth for all three apps:

- `make proto` (or `cd backend && buf generate`) → `backend/gen/` (Go server stubs).
- `cd frontend && pnpm buf:generate` → `frontend/src/gen/` (Connect-Web client).
- `cd genpos-desk && pnpm buf:generate` → `genpos-desk/src/gen/` (Connect-Web client used inside Tauri).

After editing a `.proto`, regenerate in **all three** locations or you will get type drift.

### Backend dev workflow (`backend/Makefile`)

| Target | Notes |
|---|---|
| `make -C backend run` / `build` | run / compile to `backend/bin/server` |
| `make -C backend test` | unit tests with `-race -cover` |
| `make -C backend test-integration` | run with `-tags=integration` (requires `genpos_test` DB; brought up by `make infra`) |
| `make -C backend test-coverage` | writes `coverage.out` + `coverage.html` |
| `make -C backend lint` / `fmt` / `vet` | `golangci-lint`, `gofmt + goimports`, `go vet` |
| `make -C backend sqlc` | regenerate from `internal/infrastructure/datastore/query/` |
| `make -C backend mock` | `go generate ./...` (mockgen) |
| `make -C backend generate` | sqlc + proto + mock |

To run a single test: `cd backend && go test -run Test_<Scope>_<Subject> ./path/to/pkg`.

`.air.toml` is configured for `air` live-reload (`tmp/server`).

### Backend conventions (mandatory)

`backend/.claude/rules/` contains two rules that **must** be followed:

- **`go-error-handling.md`** — all errors from services / gateways / datastores must use `pkg/errors` (a `samber/oops` wrapper). Use `errors.NotFound`, `errors.BadRequest`, `errors.Internal`, or `errors.Wrap`. Never bare `fmt.Errorf` / `errors.New` in those layers, never prefix messages with the function name (oops attaches the stack already), and keep messages short and domain-oriented — handlers translate them to HTTP statuses.
- **`go-testing.md`** — every test must (1) use the `Test_<Scope>_<Subject>` naming pattern (`Test_Integration_*`, `Test_<Type>Service_*`, `Test_<Type>Reader_*`, `Test_<Type>Writer_*`), (2) be table-driven (`t.Run` over a `map[string]struct{...}`), (3) compare structs/slices/maps with `github.com/google/go-cmp/cmp.Diff` (no `reflect.DeepEqual`, no `testify`), (4) call `t.Parallel()` at both the function and subtest level, (5) place all read-test fixtures under `database/testfixtures/`.

### Migrations (Atlas)

PostgreSQL migrations live in `backend/database/migrations/` with `atlas.sum`. Files are timestamp-prefixed (`YYYYMMDDhhmmss_*.sql`). They run automatically via the `migrate` Compose service (`backend/database/bootstrap/`) which also provisions the three DB roles from env passwords. Atlas dev/test env config is in `backend/database/atlas.hcl`.

## Frontend (`frontend/`)

TanStack Start + React 19 + Vite 8 + Tailwind v4. PowerSync via `@powersync/web` (browser, wa-sqlite). pnpm workspace. ConnectRPC client via `@connectrpc/connect-web`.

- `src/routes/` — file-based TanStack Router. `_auth.$subdomain.*` are tenant-scoped authenticated routes (the `$subdomain` param identifies the org); `_guest.*` are public (`signin`, `signup`).
- `src/features/` — feature modules: `auth`, `catalog`, `customers`, `inventory`, `reports`, `settings`, `shell`, `downloads`.
- `src/gen/` — generated Connect-Web stubs (do not edit; regenerate via `pnpm buf:generate`).

Commands (`cd frontend`):
- `pnpm dev` — Vite dev on :3032
- `pnpm build` — production build
- `pnpm start` — run `server.mjs` against the built output
- `pnpm buf:generate` — regenerate `src/gen/` from `backend/proto/`

`VITE_API_BASE_URL` (default `http://localhost:3031`) points at the backend.

## Desktop POS (`genpos-desk/`)

Tauri v2 + React 18 + Vite 5 + TS. **Different stack from the web frontend** — single-tenant, offline-first, talks to a local SQLite DB via Tauri commands and syncs to the cloud via PowerSync.

- `src-tauri/` — Rust crate (`app_lib`). `lib.rs` registers all `#[tauri::command]` handlers; features live in `src-tauri/src/features/` (`sales.rs`, `catalog.rs`, `inventory.rs`, `cashbook.rs`, `purchase_orders.rs`, `customer.rs`, `dashboard.rs`, `settings.rs`, `suppliers.rs`, `sync_credentials.rs`).
- `src-tauri/src/db/` — `r2d2 + r2d2_sqlite` connection pool with WAL + a `strip_accents` SQLite UDF for Vietnamese-friendly search. Migrations are versioned `V00N__*.sql` files under `db/migrations/`, applied via `rusqlite_migration` on startup. PRAGMAs (`foreign_keys`, `journal_mode=WAL`, `synchronous=NORMAL`) are set at connection init, **not** in migrations.
- `src-tauri/src/sync/` + `tauri-plugin-powersync` — offline sync. Credentials are set via the `sync_set_credentials` / `sync_clear_credentials` commands and the connection lifecycle is managed by `sync_connect` / `sync_disconnect`.
- `src/` — React UI. `src/app/` (App, layouts, providers, router), `src/features/` (one folder per domain, mirroring the Rust features), `src/gen/` (generated Connect-Web stubs for cloud calls).
- `src/app/router.tsx` uses `react-router-dom` (the desk app does **not** use TanStack Router — keep web and desk routing distinct).

Commands (`cd genpos-desk`):
- `pnpm tauri dev` — run desktop app in dev mode (also `make desk` from root)
- `pnpm tauri build` — produce platform installers under `src-tauri/target/release/bundle/`
- `pnpm dev` — frontend only in browser (no Tauri APIs)
- `pnpm test` / `pnpm test:watch` — Vitest
- `cargo test --manifest-path src-tauri/Cargo.toml` — Rust unit + integration tests
- `cargo clippy --manifest-path src-tauri/Cargo.toml -- -D warnings` — lint
- `pnpm buf:generate` — regenerate `src/gen/`

Prereqs: Node ≥18, pnpm ≥8, Rust ≥1.77, plus Tauri v2 system deps (https://v2.tauri.app/start/prerequisites/).

## PowerSync (sync layer)

`powersync/sync-rules.yaml` defines a single bucket per organization (`by_org`, parameterized by `token_parameters.o`). All synced tables are tenant-scoped via `org_id = bucket.org_id` (some additionally filter `deleted_at IS NULL`). When adding a synced table:

1. Add the migration in `backend/database/migrations/`.
2. Add the SELECT to `powersync/sync-rules.yaml` (and to `deploy/sync-rules.yaml` for prod).
3. The `powersync_user` role needs `SELECT` on the new table.
4. Restart the `powersync` Compose service to pick up rule changes.

The backend signs PowerSync JWTs in `pkg/auth/powersync.go`.

## Deployment

`deploy/` is the production stack: backend + frontend are pulled from GHCR, Caddy is built locally with the Cloudflare DNS plugin, Postgres + Redis + PowerSync run on the same host. Workflow: build on dev, push to GHCR, `rsync deploy/ → /opt/genpos/`, `docker compose up -d`. Full walkthrough including the three Cloudflare TLS strategies is in `docs/deploy-ubuntu.md`.
