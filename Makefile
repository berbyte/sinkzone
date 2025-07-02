# Makefile for Sinkzone

# Variables
BINARY_NAME=sinkzone
VERSION=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
COMMIT_HASH=$(shell git rev-parse --short HEAD)

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build flags
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT_HASH) -X main.date=$(BUILD_TIME) -s -w"

# Build targets
.PHONY: all build clean test deps lint docker-build docker-run help

all: clean deps test build

build:
	CGO_ENABLED=1 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) main.go

build-linux:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-linux-amd64 main.go

build-darwin:
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-darwin-amd64 main.go

build-windows:
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-windows-amd64.exe main.go

build-all: build-linux build-darwin build-windows

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME)-*

test:
	$(GOTEST) -v ./...

test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out

deps:
	$(GOMOD) tidy
	$(GOMOD) download

lint:
	golangci-lint run

install:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) main.go
	sudo cp $(BINARY_NAME) /usr/local/bin/
	sudo chmod +x /usr/local/bin/$(BINARY_NAME)

uninstall:
	sudo rm -f /usr/local/bin/$(BINARY_NAME)

docker-build:
	docker build -t $(BINARY_NAME):$(VERSION) .
	docker tag $(BINARY_NAME):$(VERSION) $(BINARY_NAME):latest

docker-run:
	docker run --rm -it --net=host $(BINARY_NAME):latest

docker-push:
	docker tag $(BINARY_NAME):$(VERSION) ghcr.io/berbyte/$(BINARY_NAME):$(VERSION)
	docker tag $(BINARY_NAME):$(VERSION) ghcr.io/berbyte/$(BINARY_NAME):latest
	docker push ghcr.io/berbyte/$(BINARY_NAME):$(VERSION)
	docker push ghcr.io/berbyte/$(BINARY_NAME):latest

release:
	goreleaser release --rm-dist

release-snapshot:
	goreleaser release --snapshot --rm-dist

help:
	@echo "Available targets:"
	@echo "  build          - Build binary for current platform"
	@echo "  build-linux    - Build binary for Linux"
	@echo "  build-darwin   - Build binary for macOS"
	@echo "  build-windows  - Build binary for Windows"
	@echo "  build-all      - Build binaries for all platforms"
	@echo "  clean          - Clean build artifacts"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage"
	@echo "  deps           - Download and tidy dependencies"
	@echo "  lint           - Run linter"
	@echo "  install        - Install binary to /usr/local/bin"
	@echo "  uninstall      - Remove binary from /usr/local/bin"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run     - Run Docker container"
	@echo "  docker-push    - Push Docker image to registry"
	@echo "  release        - Create release with goreleaser"
	@echo "  release-snapshot - Create snapshot release"
	@echo "  help           - Show this help message" 