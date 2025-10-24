.PHONY: build run test clean help

# Build the binary to bin directory
build:
	@echo "Building wordle..."
	@mkdir -p bin
	@go build -o bin/wordle ./cmd/wordle
	@echo "Build complete! Binary: bin/wordle"

# Run the game (with default small word list)
run: build
	@./bin/wordle

# Run with extended word list
run-extended: build
	@./bin/wordle -words cfg/words.txt

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
	@rm -f wordle
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
	@echo "  make build          - Build the binary to bin/wordle"
	@echo "  make run            - Build and run the game (default small word list)"
	@echo "  make run-extended   - Build and run with extended word list"
	@echo "  make test           - Run all tests"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make clean          - Remove build artifacts"
	@echo "  make deps           - Download and tidy dependencies"
	@echo "  make fmt            - Format code"
	@echo "  make lint           - Run linter"
	@echo "  make help           - Show this help message"

# Default target
.DEFAULT_GOAL := help

