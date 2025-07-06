.PHONY: build test clean install lint fmt vet help dev release test-e2e

# Build variables
BINARY_NAME=k8x
BUILD_DIR=build
VERSION?=$(shell git describe --tags --always --dirty)
COMMIT?=$(shell git rev-parse HEAD)
DATE?=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Go variables
GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/$(BUILD_DIR)
GOFILES=$(wildcard *.go)

# Build flags
LDFLAGS=-ldflags "-s -w -X github.com/shankgan/k8x/cmd.version=$(VERSION) -X github.com/shankgan/k8x/cmd.commit=$(COMMIT) -X github.com/shankgan/k8x/cmd.date=$(DATE)"

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(GOBIN)
	@go build $(LDFLAGS) -o $(GOBIN)/$(BINARY_NAME) .

test: ## Run unit tests
	@echo "Running tests..."
	@go test -v ./... -short

test-e2e: build ## Run end-to-end tests
	@echo "Running E2E tests..."
	@go test -v ./test/e2e/... -timeout 20m

test-e2e-single: build ## Run a single E2E test (usage: make test-e2e-single TEST=TestCrashLoopBackoffDiagnosis)
	@echo "Running single E2E test: $(TEST)"
	@go test -v ./test/e2e/... -run $(TEST)

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html

install: build ## Install the binary to $GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	@go install $(LDFLAGS) .

lint: ## Run golangci-lint
	@echo "Running linter..."
	@golangci-lint run

fmt: ## Run go fmt
	@echo "Formatting code..."
	@go fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

mod-tidy: ## Run go mod tidy
	@echo "Tidying modules..."
	@go mod tidy

dev: fmt vet lint mod-tidy ## Run development checks (format, vet, lint, mod-tidy)

release-test: ## Test release build with GoReleaser
	@echo "Testing release build..."
	@goreleaser release --snapshot --rm-dist

release: ## Create a release with GoReleaser
	@echo "Creating release..."
	@goreleaser release --rm-dist

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t $(BINARY_NAME):$(VERSION) .

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download

generate: ## Run go generate
	@echo "Running go generate..."
	@go generate ./...

# Development setup
setup: ## Set up development environment
	@echo "Setting up development environment..."
	@go mod download
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	@which goreleaser > /dev/null || (echo "Installing goreleaser..." && go install github.com/goreleaser/goreleaser@latest)

# Create directories
build-dir:
	@mkdir -p $(BUILD_DIR)
