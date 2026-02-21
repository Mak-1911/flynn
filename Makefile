# Flynn Makefile

.PHONY: build clean test run lint fmt help

# Variables
BINARY_NAME=flynn
DAEMON_NAME=flynnd
GO=go
GOFLAGS=-v
LDFLAGS=-s -w

# Build directories
BUILD_DIR=build
BUILD_LINUX=$(BUILD_DIR)/linux
BUILD_DARWIN=$(BUILD_DIR)/darwin
BUILD_WINDOWS=$(BUILD_DIR)/windows

help: ## Show this help message
	@echo 'Flynn - Personal AI Assistant'
	@echo ''
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build for current platform
	@echo "Building $(BINARY_NAME)..."
	@$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/flynn
	@echo "Done: $(BUILD_DIR)/$(BINARY_NAME)"

build-all: ## Build for all platforms
	@echo "Building for all platforms..."
	@mkdir -p $(BUILD_LINUX)/amd64 $(BUILD_LINUX)/arm64
	@mkdir -p $(BUILD_DARWIN)/amd64 $(BUILD_DARWIN)/arm64
	@mkdir -p $(BUILD_WINDOWS)/amd64
	@echo "Building Linux amd64..."
	@GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_LINUX)/amd64/$(BINARY_NAME) ./cmd/flynn
	@echo "Building Linux arm64..."
	@GOOS=linux GOARCH=arm64 $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_LINUX)/arm64/$(BINARY_NAME) ./cmd/flynn
	@echo "Building Darwin amd64..."
	@GOOS=darwin GOARCH=amd64 $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DARWIN)/amd64/$(BINARY_NAME) ./cmd/flynn
	@echo "Building Darwin arm64..."
	@GOOS=darwin GOARCH=arm64 $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DARWIN)/arm64/$(BINARY_NAME) ./cmd/flynn
	@echo "Building Windows amd64..."
	@GOOS=windows GOARCH=amd64 $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_WINDOWS)/amd64/$(BINARY_NAME).exe ./cmd/flynn

run: ## Run Flynn CLI
	@$(GO) run ./cmd/flynn

test: ## Run tests
	@echo "Running tests..."
	@$(GO) test -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@$(GO) test -v -coverprofile=coverage.out ./...
	@$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "Done"

fmt: ## Format Go code
	@echo "Formatting..."
	@$(GO) fmt ./...

lint: ## Run linter
	@echo "Linting..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@$(GO) mod download
	@$(GO) mod tidy

init: ## Initialize development environment
	@echo "Initializing Flynn..."
	@$(GO) mod download
	@mkdir -p ~/.flynn/models
	@mkdir -p ~/.flynn/logs
	@echo "Done. Run 'make run' to start."

daemon: ## Build and run daemon mode
	@echo "Building $(DAEMON_NAME)..."
	@$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(DAEMON_NAME) ./cmd/flynnd
	@echo "Done: $(BUILD_DIR)/$(DAEMON_NAME)"

desktop: ## Build desktop app (requires Tauri prerequisites)
	@echo "Building desktop app..."
	@cd desktop && cargo tauri build

VERSION:=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS_VERSION=-X main.version=$(VERSION)

release: clean build-all ## Create a release
	@echo "Creating release $(VERSION)..."
	@mkdir -p release
	@cd $(BUILD_LINUX)/amd64 && tar -czf ../../release/$(BINARY_NAME)-linux-amd64.tar.gz $(BINARY_NAME)
	@cd $(BUILD_LINUX)/arm64 && tar -czf ../../release/$(BINARY_NAME)-linux-arm64.tar.gz $(BINARY_NAME)
	@cd $(BUILD_DARWIN)/amd64 && tar -czf ../../release/$(BINARY_NAME)-darwin-amd64.tar.gz $(BINARY_NAME)
	@cd $(BUILD_DARWIN)/arm64 && tar -czf ../../release/$(BINARY_NAME)-darwin-arm64.tar.gz $(BINARY_NAME)
	@cd $(BUILD_WINDOWS)/amd64 && zip -q ../../release/$(BINARY_NAME)-windows-amd64.zip $(BINARY_NAME).exe
	@echo "Release files in ./release/"

.DEFAULT_GOAL := build
