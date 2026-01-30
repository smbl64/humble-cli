.PHONY: build test clean install run help

# Binary name
BINARY_NAME=humble-cli
VERSION?=dev

# Build directory
BUILD_DIR=.

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOINSTALL=$(GOCMD) install
GOMOD=$(GOCMD) mod

# Build flags
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Build the binary
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/humble-cli

build-all: ## Build for all platforms
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/humble-cli
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/humble-cli
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/humble-cli
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/humble-cli

test: ## Run tests
	$(GOTEST) -v ./...

test-coverage: ## Run tests with coverage
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out

clean: ## Clean build artifacts
	$(GOCLEAN)
	rm -f $(BUILD_DIR)/$(BINARY_NAME)
	rm -f $(BUILD_DIR)/$(BINARY_NAME)-*
	rm -f coverage.out

install: ## Install the binary to $GOPATH/bin
	$(GOINSTALL) $(LDFLAGS) ./cmd/humble-cli

run: ## Build and run
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/humble-cli
	./$(BUILD_DIR)/$(BINARY_NAME)

deps: ## Download dependencies
	$(GOMOD) download
	$(GOMOD) tidy

fmt: ## Format code
	$(GOCMD) fmt ./...

lint: ## Run linter (requires golangci-lint)
	golangci-lint run

.DEFAULT_GOAL := help
