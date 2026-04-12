.PHONY: help build run test test-coverage test-integration lint fmt vet sqlc mock clean

.DEFAULT_GOAL := help

GO := go
GOFLAGS := -v
LDFLAGS := -ldflags="-s -w"

CYAN := \033[36m
GREEN := \033[32m
YELLOW := \033[33m
RESET := \033[0m

## help: Show this help message
help:
	@echo "$(CYAN)GenPOS Backend Makefile$(RESET)"
	@echo ""
	@echo "$(GREEN)Usage:$(RESET)"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

# ==========================================
# Build Commands
# ==========================================

## build: Build the server
build:
	@echo "$(CYAN)Building server...$(RESET)"
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o bin/server ./cmd/server

## run: Run the server
run:
	@echo "$(CYAN)Running server...$(RESET)"
	$(GO) run ./cmd/server

# ==========================================
# Test Commands
# ==========================================

## test: Run all unit tests
test:
	@echo "$(CYAN)Running tests...$(RESET)"
	$(GO) test -v -race -cover ./...

## test-coverage: Run tests with coverage report
test-coverage:
	@echo "$(CYAN)Running tests with coverage...$(RESET)"
	$(GO) test -v -race -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report: coverage.html$(RESET)"

## test-integration: Run integration tests
test-integration:
	@echo "$(CYAN)Running integration tests...$(RESET)"
	$(GO) test -v -race -tags=integration ./...

# ==========================================
# Code Quality Commands
# ==========================================

## lint: Run linter
lint:
	@echo "$(CYAN)Running linter...$(RESET)"
	golangci-lint run ./...

## fmt: Format code
fmt:
	@echo "$(CYAN)Formatting code...$(RESET)"
	$(GO) fmt ./...
	goimports -w .

## vet: Run go vet
vet:
	@echo "$(CYAN)Running go vet...$(RESET)"
	$(GO) vet ./...

# ==========================================
# Code Generation Commands
# ==========================================

## sqlc: Generate sqlc code
sqlc:
	@echo "$(CYAN)Generating sqlc code...$(RESET)"
	sqlc generate

## mock: Generate mocks using mockgen
mock:
	@echo "$(CYAN)Generating mocks...$(RESET)"
	$(GO) generate ./...

## proto: Regenerate Go code from proto files
proto:
	@echo "$(CYAN)Generating proto code...$(RESET)"
	buf generate

## generate: Run all code generation
generate: sqlc proto mock

# ==========================================
# Utility Commands
# ==========================================

## clean: Clean build artifacts
clean:
	@echo "$(CYAN)Cleaning build artifacts...$(RESET)"
	rm -rf bin/
	rm -f coverage.out coverage.html

## deps: Download dependencies
deps:
	@echo "$(CYAN)Downloading dependencies...$(RESET)"
	$(GO) mod download
	$(GO) mod tidy
