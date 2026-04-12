# GenPOS Mono

Monorepo for GenPOS: PostgreSQL 17 + Redis + PowerSync + Go ConnectRPC backend + TanStack Start frontend.

## Structure

- `backend/` — Go service with ConnectRPC (proto in `backend/proto/`, generated code in `backend/gen/`)
- `frontend/` — TanStack Start (React) app
- `powersync/` — PowerSync service configuration and sync rules
- `postgres/` — PostgreSQL init scripts

## Commands

- `make infra` — start PostgreSQL, Redis, PowerSync via Docker Compose
- `make backend` — run Go backend on :8081
- `make frontend` — run frontend dev server on :3000
- `make proto` — regenerate Go code from proto files (requires buf, protoc-gen-go, protoc-gen-connect-go)

## Backend

Go module: `github.com/genpick/genpos-mono/backend`

Proto definitions live in `backend/proto/`. After editing `.proto` files, run `make proto` to regenerate.

The server starts on port 8081 (configurable via `PORT` env var).

## Frontend

Uses pnpm. Run `make frontend-install` to install dependencies.

Dev server starts on port 3000.

## Infrastructure

`docker compose up -d` starts:
- PostgreSQL 17 on :5432 (wal_level=logical for PowerSync)
- Redis 7 on :6379
- PowerSync on :8080
