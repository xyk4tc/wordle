.PHONY: build build-all build-standalone build-server build-client run run-extended test clean help

# Build all binaries
build-all: build-standalone build-server build-client

# Build standalone version (Task 1)
build-standalone:
	@echo "Building wordle (standalone)..."
	@mkdir -p bin
	@go build -o bin/wordle ./cmd/wordle
	@echo "Build complete! Binary: bin/wordle"

# Build server
build-server:
	@echo "Building wordle-server..."
	@mkdir -p bin
	@go build -o bin/wordle-server ./cmd/wordle-server
	@echo "Build complete! Binary: bin/wordle-server"

# Build client
build-client:
	@echo "Building wordle-client..."
	@mkdir -p bin
	@go build -o bin/wordle-client ./cmd/wordle-client
	@echo "Build complete! Binary: bin/wordle-client"

# Build default (standalone + server/client)
build: build-all

# Run standalone game (with default small word list)
run: build-standalone
	@./bin/wordle

# Run standalone with extended word list
run-extended: build-standalone
	@./bin/wordle -words cfg/words.txt

# Run server (with default small word list)
run-server: build-server
	@./bin/wordle-server

# Run server with extended word list
run-server-extended: build-server
	@./bin/wordle-server -words cfg/words.txt

# Run client (requires server to be running)
run-client: build-client
	@./bin/wordle-client

# Run tests
test:
	@echo "Running tests..."
	@go test ./... -v

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test ./... -cover

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f wordle wordle-server wordle-client
	@echo "Clean complete!"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Run linter
lint:
	@echo "Running linter..."
	@go vet ./...

# Show help
help:
	@echo "Available commands:"
	@echo ""
	@echo "Build commands:"
	@echo "  make build-all         - Build all binaries (standalone + server + client)"
	@echo "  make build-standalone  - Build standalone game only"
	@echo "  make build-server      - Build server only"
	@echo "  make build-client      - Build client only"
	@echo "  make build             - Same as build-all"
	@echo ""
	@echo "Run commands (Task 1 - Standalone):"
	@echo "  make run               - Run standalone game (default word list)"
	@echo "  make run-extended      - Run standalone with extended word list"
	@echo ""
	@echo "Run commands (Task 2 - Server/Client):"
	@echo "  make run-server          - Run server (default word list, port 8080)"
	@echo "  make run-server-extended - Run server with extended word list"
	@echo "  make run-client          - Run client (connects to localhost:8080)"
	@echo ""
	@echo "Other commands:"
	@echo "  make test              - Run all tests"
	@echo "  make test-coverage     - Run tests with coverage report"
	@echo "  make clean             - Remove build artifacts"
	@echo "  make deps              - Download and tidy dependencies"
	@echo "  make fmt               - Format code"
	@echo "  make lint              - Run linter"
	@echo "  make help              - Show this help message"

# Default target
.DEFAULT_GOAL := help

