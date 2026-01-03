# cool-kit Makefile
# Go project Makefile with cross-platform builds and PGO support

# =============================================================================
# Variables
# =============================================================================
BINARY_NAME := cool-kit
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GO_VERSION := $(shell go version | cut -d' ' -f3)

# Build configuration
CGO_ENABLED ?= 0
GOARM64 ?= v8.0,lse,crypto
PGO_PROFILE := default.pgo

# Linker flags
LDFLAGS := -s -w \
	-X main.version=$(VERSION) \
	-X main.commit=$(GIT_COMMIT) \
	-X main.date=$(BUILD_TIME)

# Output directories
DIST_DIR := dist
BUILD_DIR := build

# =============================================================================
# Default target
# =============================================================================
.DEFAULT_GOAL := help

.PHONY: all
all: clean lint test build

# =============================================================================
# Build targets
# =============================================================================
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	CGO_ENABLED=$(CGO_ENABLED) go build -trimpath -ldflags="$(LDFLAGS)" -o $(BINARY_NAME) .
	@echo "✅ Build complete"

.PHONY: build-dev
build-dev:
	@echo "Building $(BINARY_NAME) with race detection..."
	CGO_ENABLED=1 go build -race -ldflags="$(LDFLAGS)" -o $(BINARY_NAME)-dev .

.PHONY: build-pgo
build-pgo:
	@if [ -f $(PGO_PROFILE) ]; then \
		echo "Building with Profile-Guided Optimization..."; \
		CGO_ENABLED=$(CGO_ENABLED) go build -trimpath -pgo=$(PGO_PROFILE) \
			-ldflags="$(LDFLAGS)" \
			-o $(BINARY_NAME) .; \
		echo "✅ PGO build complete"; \
	else \
		echo "⚠️  No $(PGO_PROFILE) found. Building without PGO..."; \
		$(MAKE) build; \
	fi

# macOS universal binary (Intel + ARM64)
.PHONY: build-universal
build-universal:
	@echo "Building universal binary (macOS only)..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=$(CGO_ENABLED) \
		go build -trimpath -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-amd64 .
	GOOS=darwin GOARCH=arm64 GOARM64=$(GOARM64) CGO_ENABLED=$(CGO_ENABLED) \
		go build -trimpath -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-arm64 .
	lipo -create -output $(BINARY_NAME) $(BUILD_DIR)/$(BINARY_NAME)-amd64 $(BUILD_DIR)/$(BINARY_NAME)-arm64
	@echo "✅ Universal binary: $$(lipo -info $(BINARY_NAME))"

# =============================================================================
# Cross-platform builds
# =============================================================================
.PHONY: build-all
build-all: clean
	@echo "Building for all platforms..."
	@mkdir -p $(DIST_DIR)
	
	@echo "  → darwin/amd64"
	@GOOS=darwin GOARCH=amd64 CGO_ENABLED=$(CGO_ENABLED) \
		go build -trimpath -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 .
	
	@echo "  → darwin/arm64"
	@GOOS=darwin GOARCH=arm64 GOARM64=$(GOARM64) CGO_ENABLED=$(CGO_ENABLED) \
		go build -trimpath -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 .
	
	@echo "  → linux/amd64"
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=$(CGO_ENABLED) \
		go build -trimpath -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 .
	
	@echo "  → linux/arm64"
	@GOOS=linux GOARCH=arm64 CGO_ENABLED=$(CGO_ENABLED) \
		go build -trimpath -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 .
	
	@echo "  → windows/amd64"
	@GOOS=windows GOARCH=amd64 CGO_ENABLED=$(CGO_ENABLED) \
		go build -trimpath -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe .
	
	@echo "✅ Cross-platform builds complete:"
	@ls -lh $(DIST_DIR)/

# =============================================================================
# Dependencies
# =============================================================================
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy
	go mod verify

.PHONY: deps-update
deps-update:
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy

# =============================================================================
# Testing
# =============================================================================
.PHONY: test
test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...

.PHONY: test-short
test-short:
	@echo "Running short tests..."
	go test -short -race ./...

.PHONY: test-coverage
test-coverage: test
	@echo "Generating coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

.PHONY: benchmark
benchmark:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# =============================================================================
# Linting and formatting
# =============================================================================
.PHONY: lint
lint:
	@echo "Running linters..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --timeout=5m; \
	else \
		echo "golangci-lint not found, running go vet..."; \
		go vet ./...; \
	fi

.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	fi

.PHONY: check
check: fmt lint test

# =============================================================================
# Installation
# =============================================================================
.PHONY: install
install: build
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	sudo install -m 755 $(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	@echo "✅ Installed $(BINARY_NAME)"

.PHONY: uninstall
uninstall:
	@echo "Removing $(BINARY_NAME) from /usr/local/bin..."
	sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "✅ Uninstalled $(BINARY_NAME)"

# =============================================================================
# Development
# =============================================================================
.PHONY: dev
dev: deps fmt lint test build
	@echo "✅ Development workflow complete"

.PHONY: run
run: build
	./$(BINARY_NAME) --help

# =============================================================================
# Docker
# =============================================================================
.PHONY: docker-build
docker-build:
	docker build -t $(BINARY_NAME):$(VERSION) -t $(BINARY_NAME):latest .

.PHONY: docker-run
docker-run:
	docker run --rm $(BINARY_NAME):latest --help

# =============================================================================
# Release (GoReleaser)
# =============================================================================
.PHONY: release-check
release-check:
	goreleaser check

.PHONY: release-snapshot
release-snapshot:
	goreleaser release --snapshot --clean

.PHONY: release
release:
	goreleaser release --clean

# =============================================================================
# Cleanup
# =============================================================================
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	go clean
	rm -f $(BINARY_NAME) $(BINARY_NAME)-*
	rm -rf $(DIST_DIR) $(BUILD_DIR)
	rm -f coverage.out coverage.html
	rm -f *.prof $(PGO_PROFILE)

# =============================================================================
# Tools installation
# =============================================================================
.PHONY: install-tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/goreleaser/goreleaser/v2@latest
	@echo "✅ Development tools installed"

.PHONY: setup
setup: install-tools deps
	@echo "✅ Development environment ready"

# =============================================================================
# Information
# =============================================================================
.PHONY: info
info:
	@echo "Build Information:"
	@echo "  Binary:     $(BINARY_NAME)"
	@echo "  Version:    $(VERSION)"
	@echo "  Commit:     $(GIT_COMMIT)"
	@echo "  Go:         $(GO_VERSION)"
	@echo "  CGO:        $(CGO_ENABLED)"

.PHONY: help
help:
	@echo "$(BINARY_NAME) Makefile"
	@echo ""
	@echo "Build targets:"
	@echo "  build            Build binary"
	@echo "  build-dev        Build with race detection"
	@echo "  build-pgo        Build with Profile-Guided Optimization"
	@echo "  build-universal  Build macOS universal binary"
	@echo "  build-all        Cross-platform builds"
	@echo ""
	@echo "Development:"
	@echo "  deps             Download dependencies"
	@echo "  deps-update      Update dependencies"
	@echo "  test             Run tests with coverage"
	@echo "  lint             Run linters"
	@echo "  fmt              Format code"
	@echo "  check            Run fmt, lint, and test"
	@echo "  dev              Full dev workflow"
	@echo ""
	@echo "Installation:"
	@echo "  install          Install to /usr/local/bin"
	@echo "  uninstall        Remove from /usr/local/bin"
	@echo ""
	@echo "Release:"
	@echo "  release-check    Validate GoReleaser config"
	@echo "  release-snapshot Build snapshot release"
	@echo "  release          Create release with GoReleaser"
	@echo ""
	@echo "Other:"
	@echo "  docker-build     Build Docker image"
	@echo "  clean            Clean artifacts"
	@echo "  setup            Setup dev environment"
	@echo "  info             Show build info"
	@echo "  help             Show this help"
