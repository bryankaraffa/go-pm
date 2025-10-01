# Makefile for go-pm

.PHONY: build build-dev test clean install docs

# Build with version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_SHA ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

LDFLAGS := -X main.version=$(VERSION) -X main.gitSHA=$(GIT_SHA)

# Build the CLI with version info
build:
	go build -ldflags "$(LDFLAGS)" -o bin/go-pm ./cmd/go-pm

# Build for development (always dev version)
build-dev:
	go build -ldflags "-X github.com/bryankaraffa/go-pm.version=dev -X github.com/bryankaraffa/go-pm.gitSHA=unknown" -o bin/go-pm ./cmd/go-pm

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