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
	cd backend && go run ./cmd/server/

proto:
	cd backend && buf generate

# Frontend
frontend:
	cd frontend && pnpm dev

frontend-install:
	cd frontend && pnpm install

# Dev: run backend + frontend (requires separate terminals)
dev:
	@echo "Run in separate terminals:"
	@echo "  make infra      # start postgres, redis, powersync"
	@echo "  make backend    # start go server on :3031"
	@echo "  make frontend   # start vite dev on :3032"

clean:
	docker compose down -v
	rm -rf frontend/node_modules frontend/.output
