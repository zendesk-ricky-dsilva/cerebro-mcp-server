# Makefile for cerebro-mcp-server

# Variables
BINARY_NAME=cerebro-mcp-server
GO_FILES=$(shell find . -name "*.go" -type f)
TEST_SCRIPT=test_dependencies.sh

# Default target
.PHONY: all
all: build

# Build the Go binary
.PHONY: build
build: $(BINARY_NAME)

$(BINARY_NAME): $(GO_FILES)
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) .
	@echo "✓ Build completed successfully"

# Run tests
.PHONY: test
test: $(BINARY_NAME)
	@echo "Running tests..."
	@chmod +x $(TEST_SCRIPT)
	./$(TEST_SCRIPT)

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -f $(BINARY_NAME)
	@echo "✓ Clean completed"

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy
	@echo "✓ Dependencies installed"

# Run the server in HTTP mode
.PHONY: run-http
run-http: $(BINARY_NAME)
	@echo "Starting server in HTTP mode..."
	HTTP_MODE=true ./$(BINARY_NAME)

# Run the server in MCP mode (default)
.PHONY: run-mcp
run-mcp: $(BINARY_NAME)
	@echo "Starting server in MCP mode..."
	./$(BINARY_NAME)

# Format Go code
.PHONY: fmt
fmt:
	@echo "Formatting Go code..."
	go fmt ./...
	@echo "✓ Code formatted"

# Run Go vet
.PHONY: vet
vet:
	@echo "Running go vet..."
	go vet ./...
	@echo "✓ Vet completed"

# Run all checks (format, vet, build, test)
.PHONY: check
check: fmt vet build test
	@echo "✓ All checks passed"

# Show help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build      - Build the Go binary"
	@echo "  test       - Build and run tests"
	@echo "  clean      - Remove build artifacts"
	@echo "  deps       - Install Go dependencies"
	@echo "  run-http   - Run server in HTTP mode"
	@echo "  run-mcp    - Run server in MCP mode"
	@echo "  fmt        - Format Go code"
	@echo "  vet        - Run go vet"
	@echo "  check      - Run all checks (fmt, vet, build, test)"
	@echo "  help       - Show this help message"
