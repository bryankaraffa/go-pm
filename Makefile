# Makefile for go-pm

.PHONY: build build-dev test clean install docs

# Build with version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_SHA ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')

LDFLAGS := -s -w -X main.version=$(VERSION) -X main.gitSHA=$(GIT_SHA) -X main.buildDate=$(BUILD_DATE)

# Build the CLI with version info
build:
	go tool goreleaser build --clean --single-target

# Build for development (skip validation, i.e. dirty repo check)
build-dev:
	go tool goreleaser build --clean --single-target --skip=validate

# Build for CI (snapshot because often no tags)
build-ci:
	go tool goreleaser build --clean --single-target --snapshot

# Run tests
test:
	go test -race -coverprofile=coverage.out -covermode=atomic ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Install to GOPATH/bin
install:
	go install -ldflags "$(LDFLAGS)" ./cmd/go-pm

# Generate documentation
docs:
	go doc -all ./pkg > docs/api.md

# Run the CLI
run:
	go run ./cmd/go-pm

# Show version info
version:
	@echo "Version: $(VERSION)"
	@echo "Git SHA: $(GIT_SHA)"