.PHONY: dev build test test-integration run clean docker-up docker-down lint deps

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=sentinel

# Build the application
build:
	$(GOBUILD) -o bin/$(BINARY_NAME) ./cmd/sentinel

# Run in development mode
dev:
	LOG_LEVEL=debug $(GOCMD) run ./cmd/sentinel

# Run tests
test:
	$(GOTEST) -v -race ./...

# Run integration tests (requires docker-compose up)
test-integration:
	$(GOTEST) -v -tags=integration ./...

# Run the compiled binary
run: build
	./bin/$(BINARY_NAME)

# Clean build artifacts
clean:
	rm -rf bin/
	rm -rf /tmp/sentinel/

# Start Docker Compose stack
docker-up:
	docker-compose up -d

# Stop Docker Compose stack
docker-down:
	docker-compose down -v

# Run linter
lint:
	golangci-lint run ./...

# Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Generate mocks for testing
mocks:
	mockgen -source=internal/tools/interface.go -destination=internal/tools/mocks/mock_tool.go
	mockgen -source=internal/agent/llm.go -destination=internal/agent/mocks/mock_llm.go

# Watch for changes and rebuild (requires entr)
watch:
	find . -name "*.go" | entr -r make dev

# Help
help:
	@echo "Available targets:"
	@echo "  dev              - Run in development mode with debug logging"
	@echo "  build            - Build the binary"
	@echo "  test             - Run unit tests"
	@echo "  test-integration - Run integration tests"
	@echo "  docker-up        - Start Docker Compose stack"
	@echo "  docker-down      - Stop Docker Compose stack"
	@echo "  lint             - Run linter"
	@echo "  deps             - Download and tidy dependencies"
	@echo "  clean            - Remove build artifacts"
