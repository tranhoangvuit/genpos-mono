.PHONY: infra infra-down backend frontend proto clean

# Start infrastructure (PostgreSQL, Redis, PowerSync)
infra:
	docker compose up -d

infra-down:
	docker compose down

infra-logs:
	docker compose logs -f

# Backend
backend:
	cd backend && POWERSYNC_ENDPOINT=http://localhost:3034 go run ./cmd/server/

proto:
	cd backend && buf generate

# Frontend
frontend:
	cd frontend && pnpm dev

frontend-install:
	cd frontend && pnpm install

# Desk (Tauri app)
desk:
	cd genpos-desk && pnpm tauri dev

desk-install:
	cd genpos-desk && pnpm install

desk-proto:
	cd genpos-desk && pnpm run buf:generate

# Dev: run backend + frontend (requires separate terminals)
dev:
	@echo "Run in separate terminals:"
	@echo "  make infra      # start postgres, redis, powersync"
	@echo "  make backend    # start go server on :3031"
	@echo "  make frontend   # start vite dev on :3032"

clean:
	docker compose down -v
	rm -rf frontend/node_modules frontend/.output
