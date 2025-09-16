# Makefile for devbox CLI tool

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Binary name
BINARY_NAME=devbox
BINARY_PATH=./cmd/devbox

# Build directory
BUILD_DIR=./build

# Default target
all: clean deps build

# Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(BINARY_PATH)
	@echo "Built $(BUILD_DIR)/$(BINARY_NAME)"

# Install the binary to /usr/local/bin
install: build
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	sudo chmod +x /usr/local/bin/$(BINARY_NAME)
	@echo "Installation complete. You can now use '$(BINARY_NAME)' from anywhere."

# Build for development (current OS/arch)
dev:
	@echo "Building $(BINARY_NAME) for development..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(BINARY_PATH)
	@echo "Built $(BUILD_DIR)/$(BINARY_NAME)"

# Run tests
test:
	$(GOTEST) -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)

# Format code
fmt:
	$(GOCMD) fmt ./...

# Run linter
lint:
	golangci-lint run

# Show help
help:
	@echo "Available targets:"
	@echo "  build    - Build the binary for Linux AMD64"
	@echo "  dev      - Build the binary for current OS/arch"
	@echo "  install  - Install the binary to /usr/local/bin"
	@echo "  test     - Run tests"
	@echo "  clean    - Clean build artifacts"
	@echo "  deps     - Download and tidy dependencies"
	@echo "  fmt      - Format code"
	@echo "  lint     - Run linter"
	@echo "  help     - Show this help message"

.PHONY: all build dev install test clean deps fmt lint help