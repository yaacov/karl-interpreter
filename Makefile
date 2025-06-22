# KARL - Kubernetes Affinity Rule Language

.PHONY: build build-static test clean install lint fmt vet deps install-golangci-lint sha256

# Build variables
BINARY_NAME := karl
BUILD_DIR := ./bin
CMD_DIR := ./cmd/karl
PKG_DIR := ./pkg/karl

##@ Build

## build: Build the KARL binary
build:
	@echo "Building KARL binary..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)
	@echo "Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

## build-static: Build the KARL binary statically linked
build-static:
	@echo "Building static KARL binary..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -extldflags=-static" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(CMD_DIR)
	@echo "Static binary built: $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64"

## sha256: Generate SHA256 signature for the static binary
sha256: build-static
	@echo "Generating SHA256 signature..."
	@cd $(BUILD_DIR) && sha256sum $(BINARY_NAME)-linux-amd64 > $(BINARY_NAME)-linux-amd64.sha256
	@echo "SHA256 signature generated: $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64.sha256"
	@cat $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64.sha256

## install: Install KARL binary to GOPATH/bin
install:
	@echo "Installing KARL binary..."
	@go install $(CMD_DIR)
	@echo "KARL installed successfully"

## clean: Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@go clean
	@echo "Clean complete"

##@ Development

## deps: Download and tidy dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies updated"

## install-golangci-lint: Install golangci-lint using Go
install-golangci-lint:
	@echo "Installing golangci-lint..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "golangci-lint installed successfully"

## fmt: Format Go code
fmt:
	@echo "Formatting Go code..."
	@go fmt ./...
	@echo "Code formatted"

## vet: Run go vet
vet:
	@echo "Running go vet..."
	@go vet ./...
	@echo "Vet checks passed"

## lint: Run linting (requires golangci-lint)
lint:
	@echo "Running linter..."
	@GOBIN=$$(go env GOPATH)/bin; \
	if [ -f "$$GOBIN/golangci-lint" ]; then \
		$$GOBIN/golangci-lint run; \
	else \
		echo "golangci-lint not found in $$GOBIN. Run 'make install-golangci-lint' first."; \
		exit 1; \
	fi

##@ Testing

## test: Run all tests
test:
	@echo "Running tests..."
	@go test ./... -v
	@echo "Tests completed"

## test-coverage: Run tests with coverage report
test-coverage:
	@echo "Running tests with coverage..."
	@go test ./... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
