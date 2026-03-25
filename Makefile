.PHONY: help build test lint clean install run-analyze run-evaluate fmt vet coverage

# Variables
BINARY_NAME=kubecaps
BUILD_DIR=bin
GO=go
GOFLAGS=-v
LDFLAGS=-ldflags "-s -w"

help: ## Display this help message
	@echo "KubeCaps - Kubernetes Resource & Autoscaling Advisor"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/kubecaps
	@echo "Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

build-all: ## Build for multiple platforms
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/kubecaps
	GOOS=darwin GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/kubecaps
	GOOS=darwin GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/kubecaps
	GOOS=windows GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/kubecaps
	@echo "Multi-platform build complete"

install: build ## Install the binary to /usr/local/bin
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "Installation complete"

test: ## Run tests
	@echo "Running tests..."
	$(GO) test -v -race -coverprofile=coverage.out ./...
	@echo "Tests complete"

test-short: ## Run short tests only
	@echo "Running short tests..."
	$(GO) test -short -v ./...

test-ai: ## Run AI-specific tests
	@echo "Running AI tests..."
	$(GO) test -v ./internal/ai/...
	@echo "AI tests complete"

test-integration: ## Run integration tests with examples
	@echo "Running integration tests..."
	@echo "Testing anomaly detection..."
	$(GO) test -v ./internal/ai -run TestAnomalyDetector_RealWorldScenario
	@echo "Testing workload DNA..."
	$(GO) test -v ./internal/ai -run TestAnomalyDetector_DetectPatternAnomalies
	@echo "Integration tests complete"

test-examples: ## Run all example test scenarios
	@echo "Running example test scenarios..."
	@echo ""
	@echo "=== Test Scenario 1: Steady Workload ==="
	$(GO) test -v ./internal/ai -run "TestAnomalyDetector_DetectAnomalies/no_anomalies"
	@echo ""
	@echo "=== Test Scenario 2: Single Spike Detection ==="
	$(GO) test -v ./internal/ai -run "TestAnomalyDetector_DetectAnomalies/detect_single_spike"
	@echo ""
	@echo "=== Test Scenario 3: Multiple Spikes ==="
	$(GO) test -v ./internal/ai -run "TestAnomalyDetector_DetectAnomalies/detect_multiple"
	@echo ""
	@echo "=== Test Scenario 4: Pattern Shift ==="
	$(GO) test -v ./internal/ai -run TestAnomalyDetector_DetectPatternAnomalies
	@echo ""
	@echo "=== Test Scenario 5: Real-World Web Service ==="
	$(GO) test -v ./internal/ai -run TestAnomalyDetector_RealWorldScenario
	@echo ""
	@echo "All example scenarios complete!"

coverage: test ## Generate coverage report
	@echo "Generating coverage report..."
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

bench: ## Run benchmarks
	@echo "Running benchmarks..."
	$(GO) test -bench=. -benchmem ./...

lint: ## Run linters
	@echo "Running linters..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Install from https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run ./...

fmt: ## Format code
	@echo "Formatting code..."
	$(GO) fmt ./...
	@echo "Code formatted"

vet: ## Run go vet
	@echo "Running go vet..."
	$(GO) vet ./...

tidy: ## Tidy go modules
	@echo "Tidying go modules..."
	$(GO) mod tidy
	@echo "Modules tidied"

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

run-analyze: build ## Run analyze command (requires PROMETHEUS_URL and NAMESPACE)
	@echo "Running analyze command..."
	@$(BUILD_DIR)/$(BINARY_NAME) analyze --prometheus-url=$(PROMETHEUS_URL) --namespace=$(NAMESPACE)

run-evaluate: build ## Run evaluate command (requires PROMETHEUS_URL, NAMESPACE, and WORKLOAD)
	@echo "Running evaluate command..."
	@$(BUILD_DIR)/$(BINARY_NAME) evaluate $(WORKLOAD) --prometheus-url=$(PROMETHEUS_URL) --namespace=$(NAMESPACE)

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t kubecaps:latest .

docker-run: ## Run in Docker (requires PROMETHEUS_URL and NAMESPACE)
	@echo "Running in Docker..."
	docker run --rm -v ~/.kube:/root/.kube kubecaps:latest analyze --prometheus-url=$(PROMETHEUS_URL) --namespace=$(NAMESPACE)

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GO) mod download
	@echo "Dependencies downloaded"

verify: fmt vet lint test ## Run all verification steps

ci: verify build ## Run CI pipeline locally

dev: ## Run in development mode with hot reload (requires air)
	@which air > /dev/null || (echo "air not found. Install with: go install github.com/cosmtrek/air@latest" && exit 1)
	air

.DEFAULT_GOAL := help