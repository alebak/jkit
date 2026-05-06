.PHONY: build test vet fmt install clean help

BINARY := jkit
MODULE := github.com/alebak/jkit
BUILD_DIR := dist

# Build flags
LDFLAGS := -s -w

## build: Compile the jkit binary
build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/jkit/

## build-all: Cross-compile for all supported targets
build-all: build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64

build-linux-amd64:
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/jkit-linux-amd64 ./cmd/jkit/

build-linux-arm64:
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/jkit-linux-arm64 ./cmd/jkit/

build-darwin-amd64:
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/jkit-darwin-amd64 ./cmd/jkit/

build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/jkit-darwin-arm64 ./cmd/jkit/

## test: Run all tests
test:
	go test ./... -count=1

## test-verbose: Run all tests with verbose output
test-verbose:
	go test ./... -v -count=1

## test-cover: Run all tests with coverage report
test-cover:
	go test ./... -coverprofile=coverage.out -count=1
	go tool cover -func=coverage.out

## vet: Run go vet
vet:
	go vet ./...

## fmt: Check formatting
fmt:
	gofmt -l .

## fmt-fix: Apply formatting
fmt-fix:
	gofmt -w .

## install: Install jkit to $HOME/.local/bin
install: build
	mkdir -p $(HOME)/.local/bin
	cp $(BINARY) $(HOME)/.local/bin/
	@echo "✅ Installed to $(HOME)/.local/bin/jkit"

## clean: Remove build artifacts
clean:
	rm -f $(BINARY) coverage.out
	rm -rf $(BUILD_DIR)

## tidy: Tidy go modules
tidy:
	go mod tidy

## help: Show this help
help:
	@grep '^##' Makefile | sed 's/^## //'
