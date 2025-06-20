# Makefile for Gosh - A modern shell written in Go
# This Makefile provides targets for building, testing, and managing the gosh project

# Variables
BINARY_NAME=gosh
MAIN_PATH=cmd/main.go
BUILD_DIR=build
INSTALL_PATH=/usr/local/bin
CONFIG_DIR=$(HOME)/.config/gosh
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Go build flags
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildTime=$(BUILD_TIME)"
BUILD_FLAGS=-trimpath $(LDFLAGS)

# Default target
.PHONY: all
all: build

# Help target
.PHONY: help
help:
	@echo "Gosh Build System"
	@echo "=================="
	@echo ""
	@echo "Available targets:"
	@echo "  build          Build the gosh binary"
	@echo "  run            Build and run gosh from the repository"
	@echo "  run-debug      Build and run gosh in debug mode"
	@echo "  install        Build and install gosh to $(INSTALL_PATH)"
	@echo "  uninstall      Remove gosh from $(INSTALL_PATH)"
	@echo "  test           Run all tests"
	@echo "  test-unit      Run unit tests"
	@echo "  test-integration Run integration tests"
	@echo "  bench          Run benchmarks"
	@echo "  clean          Clean build artifacts"
	@echo "  fmt            Format Go code"
	@echo "  lint           Run linters"
	@echo "  vet            Run go vet"
	@echo "  deps           Download dependencies"
	@echo "  deps-update    Update dependencies"
	@echo "  setup          Run setup script"
	@echo "  docs           Generate documentation"
	@echo "  release        Build release binaries"
	@echo "  docker         Build Docker image"
	@echo "  docker-run     Run gosh in Docker container"
	@echo "  version        Show version information"
	@echo "  help           Show this help message"

# Build targets
.PHONY: build
build: deps
	@echo "Building gosh..."
	@mkdir -p $(BUILD_DIR)
	go build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

.PHONY: build-debug
build-debug: deps
	@echo "Building gosh with debug symbols..."
	@mkdir -p $(BUILD_DIR)
	go build -gcflags="all=-N -l" -o $(BUILD_DIR)/$(BINARY_NAME)-debug $(MAIN_PATH)
	@echo "Debug build complete: $(BUILD_DIR)/$(BINARY_NAME)-debug"

.PHONY: build-race
build-race: deps
	@echo "Building gosh with race detection..."
	@mkdir -p $(BUILD_DIR)
	go build -race $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-race $(MAIN_PATH)
	@echo "Race build complete: $(BUILD_DIR)/$(BINARY_NAME)-race"

# Installation targets
.PHONY: install
install: build
	@echo "Installing gosh to $(INSTALL_PATH)..."
	@if [ -w "$(dir $(INSTALL_PATH))" ]; then \
		cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_PATH); \
		chmod +x $(INSTALL_PATH); \
	else \
		echo "Installing to $(INSTALL_PATH) requires sudo..."; \
		sudo cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_PATH); \
		sudo chmod +x $(INSTALL_PATH); \
	fi
	@echo "Installation complete"

.PHONY: uninstall
uninstall:
	@echo "Removing gosh from $(INSTALL_PATH)..."
	@if [ -f "$(INSTALL_PATH)" ]; then \
		if [ -w "$(dir $(INSTALL_PATH))" ]; then \
			rm $(INSTALL_PATH); \
		else \
			sudo rm $(INSTALL_PATH); \
		fi; \
		echo "Uninstallation complete"; \
	else \
		echo "Gosh not found at $(INSTALL_PATH)"; \
	fi

# Testing targets
.PHONY: test
test: test-unit test-integration

.PHONY: test-unit
test-unit:
	@echo "Running unit tests..."
	go test -v ./internal/...

.PHONY: test-integration
test-integration:
	@echo "Running integration tests..."
	go test -v -tags=integration ./...

.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./internal/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: bench
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./internal/...

# Code quality targets
.PHONY: fmt
fmt:
	@echo "Formatting Go code..."
	go fmt ./...
	@echo "Formatting complete"

.PHONY: lint
lint:
	@echo "Running linters..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found, install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

.PHONY: vet
vet:
	@echo "Running go vet..."
	go vet ./...

.PHONY: check
check: fmt vet lint test
	@echo "All checks passed"

# Dependency management
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

.PHONY: deps-update
deps-update:
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy

.PHONY: deps-vendor
deps-vendor:
	@echo "Vendoring dependencies..."
	go mod vendor

# Setup and configuration
.PHONY: setup
setup:
	@echo "Running setup script..."
	./setup.sh

.PHONY: config-sample
config-sample:
	@echo "Creating sample configuration files..."
	@mkdir -p $(CONFIG_DIR)
	@if [ ! -f "$(HOME)/.goshrc" ]; then \
		cp docs/sample.goshrc $(HOME)/.goshrc; \
		echo "Created ~/.goshrc"; \
	fi
	@if [ ! -f "$(HOME)/.gosh_profile" ]; then \
		cp docs/sample.gosh_profile $(HOME)/.gosh_profile; \
		echo "Created ~/.gosh_profile"; \
	fi

# Documentation
.PHONY: docs
docs:
	@echo "Generating documentation..."
	@if command -v godoc >/dev/null 2>&1; then \
		echo "Starting godoc server at http://localhost:6060"; \
		godoc -http=:6060; \
	else \
		echo "godoc not found, install with: go install golang.org/x/tools/cmd/godoc@latest"; \
	fi

# Release targets
.PHONY: release
release: clean
	@echo "Building release binaries..."
	@mkdir -p $(BUILD_DIR)/release

	# Linux AMD64
	GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)

	# Linux ARM64
	GOOS=linux GOARCH=arm64 go build $(BUILD_FLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)

	# macOS AMD64
	GOOS=darwin GOARCH=amd64 go build $(BUILD_FLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)

	# macOS ARM64 (Apple Silicon)
	GOOS=darwin GOARCH=arm64 go build $(BUILD_FLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)

	# Windows AMD64
	GOOS=windows GOARCH=amd64 go build $(BUILD_FLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)

	@echo "Release binaries built in $(BUILD_DIR)/release/"

.PHONY: package
package: release
	@echo "Creating release packages..."
	@cd $(BUILD_DIR)/release && \
	for binary in $(BINARY_NAME)-*; do \
		if [[ $$binary == *.exe ]]; then \
			zip $$binary.zip $$binary; \
		else \
			tar -czf $$binary.tar.gz $$binary; \
		fi; \
	done
	@echo "Release packages created in $(BUILD_DIR)/release/"

# Docker targets
.PHONY: docker
docker:
	@echo "Building Docker image..."
	docker build -t gosh:$(VERSION) .
	docker tag gosh:$(VERSION) gosh:latest

.PHONY: docker-run
docker-run:
	@echo "Running gosh in Docker container..."
	docker run -it --rm gosh:latest

# Utility targets
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	go clean -cache -testcache -modcache
	@echo "Clean complete"

.PHONY: version
version:
	@echo "Version: $(VERSION)"
	@echo "Commit:  $(COMMIT)"
	@echo "Built:   $(BUILD_TIME)"

.PHONY: info
info:
	@echo "Project Information"
	@echo "==================="
	@echo "Binary:     $(BINARY_NAME)"
	@echo "Version:    $(VERSION)"
	@echo "Commit:     $(COMMIT)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Go Version: $(shell go version)"
	@echo "Build Dir:  $(BUILD_DIR)"
	@echo "Install:    $(INSTALL_PATH)"
	@echo "Config:     $(CONFIG_DIR)"

# Development targets
.PHONY: run
run: build
	@echo "Starting gosh..."
	./$(BUILD_DIR)/$(BINARY_NAME)

.PHONY: run-debug
run-debug: build
	@echo "Starting gosh in debug mode..."
	./$(BUILD_DIR)/$(BINARY_NAME) --debug

.PHONY: dev
dev: build
	@echo "Starting development build..."
	./$(BUILD_DIR)/$(BINARY_NAME) --debug

.PHONY: watch
watch:
	@echo "Watching for changes..."
	@if command -v fswatch >/dev/null 2>&1; then \
		fswatch -o . | xargs -n1 -I{} make build; \
	elif command -v inotifywait >/dev/null 2>&1; then \
		while inotifywait -r -e modify .; do make build; done; \
	else \
		echo "Install fswatch or inotify-tools for file watching"; \
	fi

# Note: Build directory is created automatically by build targets
