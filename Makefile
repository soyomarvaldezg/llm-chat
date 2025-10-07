.PHONY: all build clean install uninstall test run help deps tidy build-all

# Binary name
BINARY_NAME=llm-chat

# Build directory
BUILD_DIR=build

# Go path
GOPATH=$(shell go env GOPATH)
INSTALL_PATH=$(GOPATH)/bin

# Version info
VERSION?=0.1.0
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildDate=$(BUILD_DATE)"

# Default target
all: build

## help: Show this help message
help:
	@echo "Available commands:"
	@echo ""
	@echo "  make build       - Build the binary to ./build/"
	@echo "  make install     - Install to GOPATH/bin"
	@echo "  make uninstall   - Remove from GOPATH/bin"
	@echo "  make clean       - Remove build artifacts"
	@echo "  make test        - Run tests"
	@echo "  make run         - Build and run"
	@echo "  make deps        - Download dependencies"
	@echo "  make tidy        - Tidy dependencies"
	@echo "  make build-all   - Build for multiple platforms"

## build: Build the binary to ./build/
build:
	@echo "Building $(BINARY_NAME) v$(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/llm-chat
	@echo "✅ Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

## install: Install to GOPATH/bin
install:
	@echo "Building and installing $(BINARY_NAME) to $(INSTALL_PATH)..."
	go build $(LDFLAGS) -o $(INSTALL_PATH)/$(BINARY_NAME) ./cmd/llm-chat
	@echo "✅ $(BINARY_NAME) installed successfully!"
	@echo "   Run '$(BINARY_NAME) --help' to get started."
	@echo "   Make sure $(INSTALL_PATH) is in your PATH"

## uninstall: Remove from GOPATH/bin
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	@rm -f "$(INSTALL_PATH)/$(BINARY_NAME)"
	@echo "✅ $(BINARY_NAME) has been removed."

## clean: Remove build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@echo "✅ Clean complete"

## test: Run tests
test:
	@echo "Running tests..."
	go test -v ./...

## run: Build and run
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME)

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	go get github.com/ollama/ollama/api
	go get github.com/sashabaranov/go-openai
	go get github.com/google/generative-ai-go/genai
	go get google.golang.org/api/option
	go get github.com/spf13/cobra
	go get github.com/fatih/color
	@echo "✅ Dependencies downloaded"

## tidy: Tidy and verify dependencies
tidy:
	@echo "Tidying dependencies..."
	go mod tidy
	go mod verify
	@echo "✅ Dependencies tidied"

## build-all: Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/llm-chat
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/llm-chat
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/llm-chat
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/llm-chat
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/llm-chat
	@echo "✅ Multi-platform build complete"
