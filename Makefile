# Sukuk POC Backend - Makefile

.PHONY: help build run test test-coverage lint clean swag docs

# Default target
.DEFAULT_GOAL := help

# Application name
APP_NAME := sukuk-poc-api
BINARY_DIR := bin

help: ## Show this help message
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the application
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BINARY_DIR)
	@go build -o $(BINARY_DIR)/$(APP_NAME) main.go
	@echo "Built $(APP_NAME) in $(BINARY_DIR)/"

run: ## Run the application
	@echo "Running $(APP_NAME)..."
	@go run main.go

test: ## Run all tests
	@echo "Running tests..."
	@go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -cover ./...

test-coverage-profile: ## Generate coverage profile
	@echo "Generating coverage profile..."
	@mkdir -p coverage
	@go test -coverprofile=coverage/coverage.out -v ./...
	@go tool cover -func=coverage/coverage.out
	@echo "Coverage profile saved to coverage/coverage.out"

lint: ## Run linter (if available)
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, running go vet..."; \
		go vet ./...; \
	fi

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf $(BINARY_DIR)
	@rm -rf coverage
	@go clean

# Database commands
migrate: ## Run database migrations
	@echo "Running database migrations..."
	@go run cmd/migrate/main.go

seed: ## Seed database with sample data
	@echo "Seeding database..."
	@go run cmd/seed/main.go

# Documentation commands
swag: ## Generate Swagger documentation
	@echo "Generating Swagger documentation..."
	@$(HOME)/go/bin/swag init -g main.go -o docs --parseDependency --parseInternal
	@echo "Swagger documentation generated in docs/"

docs: swag ## Generate and serve documentation locally
	@echo "Documentation available at http://localhost:8080/swagger/index.html"