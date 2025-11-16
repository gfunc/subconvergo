# Makefile for subconvergo

.PHONY: help build test test-unit test-integration test-all test-docker clean coverage lint run dev install deps docker-build docker-test docker-run

# Variables
BINARY_NAME=subconvergo
GO=go
GOTEST=$(GO) test
GOBUILD=$(GO) build
GOCLEAN=$(GO) clean
DOCKER=docker
DOCKER_COMPOSE=docker-compose

# Build variables
VERSION?=dev
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build targets
build: ## Build the application
	@echo "Building $(BINARY_NAME)..."
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) .

build-all: ## Build for all platforms
	@echo "Building for multiple platforms..."
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-linux-amd64 .
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-linux-arm64 .
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-darwin-arm64 .
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-windows-amd64.exe .

# Test targets
test: ## Run smoke tests (integration/API tests with Docker)
	@echo "Running smoke tests..."
	python -m tests.smoke

test-unit: ## Run unit tests only
	@echo "Running unit tests..."
	./tests/run-tests.sh unit

test-all: test-unit test ## Run unit tests and smoke tests
	@echo "All tests completed"

# Coverage targets
coverage: ## Generate coverage report
	@echo "Generating coverage report..."
	./tests/run-tests.sh coverage

coverage-view: coverage ## Open coverage report in browser
	@echo "Opening coverage report..."
	@which xdg-open > /dev/null && xdg-open coverage/coverage.html || \
	 which open > /dev/null && open coverage/coverage.html || \
	 echo "Please open coverage/coverage.html manually"

# Lint and format
lint: ## Run linter
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Run: make install-tools"; exit 1)
	golangci-lint run ./...

fmt: ## Format code
	@echo "Formatting code..."
	$(GO) fmt ./...
	@which goimports > /dev/null && goimports -w . || echo "goimports not installed"

vet: ## Run go vet
	@echo "Running go vet..."
	$(GO) vet ./...

# Dependencies
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GO) mod download

deps-update: ## Update dependencies
	@echo "Updating dependencies..."
	$(GO) get -u ./...
	$(GO) mod tidy

deps-verify: ## Verify dependencies
	@echo "Verifying dependencies..."
	$(GO) mod verify

# Development
dev: ## Run in development mode
	@echo "Running in development mode..."
	$(GO) run . -f pref.toml

run: build ## Build and run
	@echo "Running $(BINARY_NAME)..."
	./$(BINARY_NAME)

watch: ## Watch for changes and rebuild (requires entr)
	@echo "Watching for changes..."
	@which entr > /dev/null || (echo "entr not installed. Install with: brew install entr"; exit 1)
	find . -name '*.go' | entr -r make run

# Docker targets
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	$(DOCKER) build -t $(BINARY_NAME):$(VERSION) .
	$(DOCKER) tag $(BINARY_NAME):$(VERSION) $(BINARY_NAME):latest

docker-run: docker-build ## Run Docker container
	@echo "Running Docker container..."
	$(DOCKER) run --rm -p 25500:25500 \
		-v $(PWD)/base:/app/base \
		-v $(PWD)/pref.toml:/app/pref.toml:ro \
		$(BINARY_NAME):latest

# Benchmark
bench: ## Run benchmarks
	@echo "Running benchmarks..."
	./tests/run-tests.sh bench

# Security
security-scan: ## Run security scan
	@echo "Running security scan..."
	@which gosec > /dev/null || (echo "gosec not installed. Run: make install-tools"; exit 1)
	gosec ./...

vulnerability-check: ## Check for vulnerabilities
	@echo "Checking for vulnerabilities..."
	$(GO) list -json -m all | nancy sleuth

# Clean
clean: ## Clean build artifacts
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME)-*
	rm -rf coverage/
	rm -f test-results.txt
	rm -f docker-test-logs.txt
	rm -rf vendor/

clean-all: clean ## Clean everything including caches
	@echo "Cleaning all..."
	$(GO) clean -cache -testcache -modcache

# Install tools
install-tools: ## Install development tools
	@echo "Installing development tools..."
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GO) install github.com/securego/gosec/v2/cmd/gosec@latest
	$(GO) install golang.org/x/tools/cmd/goimports@latest
	$(GO) install golang.org/x/perf/cmd/benchstat@latest
	@echo "Tools installed successfully"

# Install
install: build ## Install binary to GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	$(GO) install $(LDFLAGS) .

# CI/CD
ci: deps lint vet test-unit test ## Run CI pipeline
	@echo "CI pipeline completed"

# Version
version: ## Show version information
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Git Commit: $(GIT_COMMIT)"

.DEFAULT_GOAL := help
