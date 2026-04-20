# PHONY targets (these don't represent files)
.PHONY: all build build-linux run lint clean install-lint test docker-build docker-run

# Go parameters
BINARY_NAME=basilik
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
