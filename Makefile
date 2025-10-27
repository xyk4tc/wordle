.PHONY: build build-all build-server build-client run run-client run-server run-server-extended test clean help

# Build all binaries
build-all: build-server build-client

# Build server
build-server:
	@echo "Building wordle-server..."
	@mkdir -p bin
	@go build -o bin/wordle-server ./cmd/wordle-server
	@echo "Build complete! Binary: bin/wordle-server"

# Build client (includes all 3 modes: offline, single-player, multi-player)
build-client:
	@echo "Building wordle-client (unified client with all modes)..."
	@mkdir -p bin
	@go build -o bin/wordle-client ./cmd/wordle-client
	@echo "Build complete! Binary: bin/wordle-client"

# Build default (server + client)
build: build-all

# Run client (interactive mode selection)
run: build-client
	@./bin/wordle-client

# Run client in offline mode (no server required)
run-offline: build-client
	@./bin/wordle-client -mode offline

# Run client with extended word list in offline mode
run-offline-extended: build-client
	@./bin/wordle-client -mode offline -words cfg/words.txt

# Run server (with default small word list)
run-server: build-server
	@./bin/wordle-server

# Run server with extended word list
run-server-extended: build-server
	@./bin/wordle-server -words cfg/words.txt

# Run client (same as 'make run')
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
	@echo "  make build-all         - Build all binaries (server + client)"
	@echo "  make build-server      - Build server only"
	@echo "  make build-client      - Build unified client (all 3 modes)"
	@echo "  make build             - Same as build-all"
	@echo ""
	@echo "Run commands (Client - Unified):"
	@echo "  make run                  - Run client (interactive mode selection)"
	@echo "  make run-client           - Same as 'make run'"
	@echo "  make run-offline          - Run in offline mode (no server)"
	@echo "  make run-offline-extended - Run offline with extended word list"
	@echo ""
	@echo "Run commands (Server):"
	@echo "  make run-server          - Run server (default word list, port 8080)"
	@echo "  make run-server-extended - Run server with extended word list"
	@echo ""
	@echo "Client supports 3 modes:"
	@echo "  0. Offline      - Task 1: Standalone, no server required"
	@echo "  1. Single-Player - Task 2: Online single-player"
	@echo "  2. Multi-Player  - Task 4: Online multiplayer racing"
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

