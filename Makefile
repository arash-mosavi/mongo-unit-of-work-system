# Unit of Work Template Project Makefile
# Production-ready development workflow

.PHONY: help build test test-race test-cover bench clean lint fmt vet deps tidy run-example

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build targets
build: ## Build the project
	@echo "Building project..."
	@go build -v ./...

# Test targets
test: ## Run all tests
	@echo "Running tests..."
	@go test -v ./...

test-race: ## Run tests with race detection
	@echo "Running tests with race detection..."
	@go test -race -v ./...

test-cover: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-integration: ## Run integration tests (requires PostgreSQL)
	@echo "Running integration tests..."
	@go test -tags=integration -v ./pkg/postgres

bench: ## Run benchmarks
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem ./...

# Code quality targets
lint: ## Run golangci-lint
	@echo "Running linter..."
	@golangci-lint run

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...
	@goimports -w .

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

# Dependency management
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download

tidy: ## Tidy dependencies
	@echo "Tidying dependencies..."
	@go mod tidy

# Example targets
run-example: ## Run usage example
	@echo "Running usage example..."
	@go run examples/usage.go

# Development workflow
dev-setup: deps ## Setup development environment
	@echo "Setting up development environment..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/tools/cmd/goimports@latest

# CI/CD targets
ci: fmt vet lint test test-race ## Run all CI checks

# Database targets
db-up: ## Start PostgreSQL container for development
	@echo "Starting PostgreSQL container..."
	@docker run --name unit-of-work-postgres \
		-e POSTGRES_DB=unit_of_work_dev \
		-e POSTGRES_USER=postgres \
		-e POSTGRES_PASSWORD=password \
		-p 5432:5432 \
		-d postgres:15-alpine

db-down: ## Stop PostgreSQL container
	@echo "Stopping PostgreSQL container..."
	@docker stop unit-of-work-postgres || true
	@docker rm unit-of-work-postgres || true

db-reset: db-down db-up ## Reset PostgreSQL container

# Documentation targets
docs: ## Generate documentation
	@echo "Generating documentation..."
	@go doc -all ./pkg/persistence > docs/persistence.md
	@go doc -all ./pkg/postgres > docs/postgres.md
	@go doc -all ./pkg/identifier > docs/identifier.md

# Cleanup targets
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@go clean -cache -testcache -modcache
	@rm -f coverage.out coverage.html

# Release targets
tag: ## Create a new git tag (usage: make tag VERSION=v1.0.0)
	@echo "Creating tag $(VERSION)..."
	@git tag -a $(VERSION) -m "Release $(VERSION)"
	@git push origin $(VERSION)

# Performance targets
profile-cpu: ## Run CPU profiling
	@echo "Running CPU profiling..."
	@go test -cpuprofile=cpu.prof -bench=. ./pkg/postgres
	@go tool pprof cpu.prof

profile-mem: ## Run memory profiling
	@echo "Running memory profiling..."
	@go test -memprofile=mem.prof -bench=. ./pkg/postgres
	@go tool pprof mem.prof

# Security targets
security: ## Run security checks
	@echo "Running security checks..."
	@gosec ./...

# Full workflow
all: clean deps fmt vet lint test test-race bench ## Run complete workflow
