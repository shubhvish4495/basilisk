# PHONY targets (these don't represent files)
.PHONY: all build build-linux run lint clean install-lint test docker-build docker-run install-migrate create-migration migrate-up migrate-down

# migration directory variable
MIGRATIONS_DIR=db/migrations

# Go parameters
BINARY_NAME=basilisk
BUILD_DIR=bin
GOLANGCI_LINT_URL=github.com/golangci/golangci-lint/cmd/golangci-lint@v1.63.4
SRC=$(shell find . -type f -name '*.go')
# Build Docker/Podman image
CONTAINER_ENGINE := $(shell command -v docker 2>/dev/null || command -v podman 2>/dev/null)
ifeq (,$(CONTAINER_ENGINE))
    $(error "Neither Docker nor Podman is installed")
endif

# Check if golangci-lint is installed
GOLANGCI_LINT=$(shell command -v golangci-lint 2>/dev/null)
MIGRATE=$(shell command -v migrate 2>/dev/null)
MIGRATE_URL=github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Default target
all: build

# Build the Go application
build: clean lint $(SRC)
	@echo "+ $@"
	@mkdir -p $(BUILD_DIR)
	@echo "    🔨 Building the binary..."
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd

# Build the Go application for Linux
build-linux: clean lint $(SRC)
	@echo "+ $@"
	@mkdir -p $(BUILD_DIR)
	@echo "    🔨 Building the binary for Linux..."
	@GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux ./cmd

# Run the application
run: build
	@echo "🚀 Running the application..."
	@$(BUILD_DIR)/$(BINARY_NAME)

# Install golangci-lint if not installed
install-lint:
	@echo "+ $@"
	@if [ -z "$(GOLANGCI_LINT)" ]; then \
		echo "    ⚙️  Installing golangci-lint..."; \
		go install $(GOLANGCI_LINT_URL); \
	else \
		echo "    ✅ golangci-lint is already installed"; \
	fi

# Run linting only on the current directory (excluding ~/go/pkg/mod)
lint: install-lint
	@echo "+ $@"
	@echo "    🔍 Running linter in current directory only..."
	@golangci-lint run --tests=false
	@echo "    ✅ Linter passed"

# Run tests
test:
	@echo "+ $@"
	@echo "    🧪 Running tests..."
	@go test -race -cover -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html
	@echo "    ✅ Tests passed"

# Clean build artifacts
clean:
	@echo "+ $@"
	@echo "    🗑️  Cleaning up..."
	@rm -rf $(BUILD_DIR)

docker-build:
	@echo "+ $@"
	@echo "    🐳 Building Docker image..."
	$(CONTAINER_ENGINE) build -t $(BINARY_NAME) .

# Run Docker container
docker-run: docker-build
	@echo "+ $@"
	@echo "    🐳 Running Docker container..."
	$(CONTAINER_ENGINE) run -it $(BINARY_NAME)

# Database URL for migrations
DB_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL_MODE)

# Install golang-migrate if not installed
install-migrate:
	@echo "+ $@"
	@if [ -z "$(MIGRATE)" ]; then \
		echo "    ⚙️  Installing golang-migrate..."; \
		go install -tags 'postgres' $(MIGRATE_URL); \
	else \
		echo "    ✅ golang-migrate is already installed"; \
	fi

# Create a new migration (usage: make create-migration name=<migration_name>)
create-migration: install-migrate
	@echo "+ $@"
	@if [ -z "$(name)" ]; then \
		echo "    ❌ Error: please provide a migration name, e.g. make create-migration name=create_users_table"; \
		exit 1; \
	fi
	@mkdir -p $(MIGRATIONS_DIR)
	@migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)
	@echo "    ✅ Migration files created"

# Run database migrations up
migrate-up: install-migrate
	@echo "+ $@"
	@echo "    ⬆️  Running migrations up..."
	@migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up
	@echo "    ✅ Migrations applied"

# Roll back database migrations
migrate-down: install-migrate
	@echo "+ $@"
	@echo "    ⬇️  Running migrations down..."
	@migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down
	@echo "    ✅ Migrations rolled back"
