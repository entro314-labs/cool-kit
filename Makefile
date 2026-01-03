.PHONY: build build-all clean install test fmt lint help

BINARY_NAME=cool-kit
VERSION?=1.0.0
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE?=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS=-ldflags "-X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${BUILD_DATE} -s -w"

.DEFAULT_GOAL := help

help:
	@echo "Coolify Deployer - Build Commands"
	@echo ""
	@echo "Usage: make <target>"
	@echo ""
	@echo "Targets:"
	@echo "  build      - Build binary for current platform"
	@echo "  build-all  - Build binaries for all platforms"
	@echo "  install    - Install binary to /usr/local/bin"
	@echo "  clean      - Remove build artifacts"
	@echo "  test       - Run tests"
	@echo "  fmt        - Format code"
	@echo "  run        - Build and run"
	@echo "  dev        - Run without building"

build:
	@echo "Building ${BINARY_NAME}..."
	go build ${LDFLAGS} -o ${BINARY_NAME} main.go
	@echo "✓ Build complete"

build-all: clean
	@echo "Building for all platforms..."
	@mkdir -p dist
	@GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o dist/${BINARY_NAME}-linux-amd64 main.go
	@GOOS=linux GOARCH=arm64 go build ${LDFLAGS} -o dist/${BINARY_NAME}-linux-arm64 main.go
	@GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -o dist/${BINARY_NAME}-darwin-amd64 main.go
	@GOOS=darwin GOARCH=arm64 go build ${LDFLAGS} -o dist/${BINARY_NAME}-darwin-arm64 main.go
	@echo "✓ All builds complete"

install: build
	@echo "Installing to /usr/local/bin..."
	@sudo mv ${BINARY_NAME} /usr/local/bin/
	@echo "✓ Installed successfully"

clean:
	@rm -f ${BINARY_NAME}
	@rm -rf dist/

test:
	go test -v ./...

fmt:
	go fmt ./...

run: build
	./${BINARY_NAME}

dev:
	go run main.go
