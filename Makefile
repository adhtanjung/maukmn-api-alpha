.PHONY: help dev run build migrate migrate-down migrate-status migrate-create test clean deps

# Load .env file if it exists
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# Default target
help:
	@echo "Maukemana Backend - Available commands:"
	@echo ""
	@echo "  make dev            - Run with hot reload (air)"
	@echo "  make run            - Run without hot reload"
	@echo "  make build          - Build the binary"
	@echo "  make migrate        - Run database migrations (up)"
	@echo "  make migrate-down   - Rollback last migration"
	@echo "  make migrate-status - Show migration status"
	@echo "  make migrate-create name=<name> - Create new migration"
	@echo "  make test           - Run tests"
	@echo "  make deps           - Install dependencies"
	@echo "  make clean          - Clean build artifacts"
	@echo ""

# Run with hot reload using air
dev:
	@echo "🚀 Starting server with hot reload..."
	@$$(go env GOPATH)/bin/air

# Run without hot reload
run:
	@echo "🚀 Starting server..."
	@go run cmd/server/main.go

# Build the binary
build:
	@echo "📦 Building..."
	@go build -o bin/server cmd/server/main.go
	@echo "✅ Built: bin/server"

# Database migrations using goose CLI
migrate:
	@echo "🔄 Running migrations..."
	@goose -dir migrations postgres "$(DATABASE_URL)" up

migrate-down:
	@echo "⬇️  Rolling back last migration..."
	@goose -dir migrations postgres "$(DATABASE_URL)" down

migrate-status:
	@echo "📋 Migration status:"
	@goose -dir migrations postgres "$(DATABASE_URL)" status

migrate-reset:
	@echo "⚠️  Resetting database..."
	@goose -dir migrations postgres "$(DATABASE_URL)" reset

migrate-create:
	@echo "📝 Creating new migration: $(name)"
	@goose -dir migrations create $(name) sql

# Run tests
test:
	@echo "🧪 Running tests..."
	@go test -v ./...

# Install dependencies
deps:
	@echo "📥 Installing dependencies..."
	@go mod download
	@go mod tidy
	@echo "✅ Dependencies installed"

# Install air for hot reload (if not installed)
install-air:
	@echo "📥 Installing air..."
	@go install github.com/air-verse/air@latest
	@echo "✅ Air installed"

# Clean build artifacts
clean:
	@echo "🧹 Cleaning..."
	@rm -rf bin/ tmp/
	@echo "✅ Cleaned"

# Format code
fmt:
	@echo "📝 Formatting code..."
	@go fmt ./...

# Lint code
lint:
	@echo "🔍 Linting..."
	@golangci-lint run

# Check health endpoint
health:
	@curl -s http://localhost:3001/health | jq .
