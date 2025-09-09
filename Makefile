# Cart Service Makefile

# Variables
APP_NAME=cart-service
DOCKER_IMAGE=aioutlet/$(APP_NAME)
VERSION?=latest
DOCKER_COMPOSE_FILE=docker-compose.yml

# Go related variables
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=$(APP_NAME)
BINARY_PATH=./cmd/server

# Default target
.PHONY: all
all: test build

# Help
.PHONY: help
help:
	@echo "Available commands:"
	@echo "  build          - Build the application"
	@echo "  run            - Run the application locally"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage"
	@echo "  clean          - Clean build artifacts"
	@echo "  deps           - Download dependencies"
	@echo "  tidy           - Tidy and verify dependencies"
	@echo "  lint           - Run linter"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run     - Run with Docker Compose"
	@echo "  docker-stop    - Stop Docker Compose"
	@echo "  docker-clean   - Remove Docker containers and images"

# Build the application
.PHONY: build
build:
	$(GOBUILD) -o $(BINARY_NAME) -v $(BINARY_PATH)

# Run the application
.PHONY: run
run:
	$(GOCMD) run $(BINARY_PATH)/main.go

# Test the application
.PHONY: test
test:
	$(GOTEST) -v ./...

# Test with coverage
.PHONY: test-coverage
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
.PHONY: clean
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

# Download dependencies
.PHONY: deps
deps:
	$(GOMOD) download

# Tidy dependencies
.PHONY: tidy
tidy:
	$(GOMOD) tidy
	$(GOMOD) verify

# Run linter (requires golangci-lint to be installed)
.PHONY: lint
lint:
	golangci-lint run

# Install linter
.PHONY: install-lint
install-lint:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.54.2

# Format code
.PHONY: fmt
fmt:
	$(GOCMD) fmt ./...

# Generate swagger docs (requires swag to be installed)
.PHONY: swagger
swagger:
	swag init -g cmd/server/main.go -o docs

# Install swagger generator
.PHONY: install-swagger
install-swagger:
	$(GOGET) github.com/swaggo/swag/cmd/swag

# Docker commands
.PHONY: docker-build
docker-build:
	docker build -t $(DOCKER_IMAGE):$(VERSION) .

.PHONY: docker-run
docker-run:
	docker-compose -f $(DOCKER_COMPOSE_FILE) up -d

.PHONY: docker-stop
docker-stop:
	docker-compose -f $(DOCKER_COMPOSE_FILE) down

.PHONY: docker-logs
docker-logs:
	docker-compose -f $(DOCKER_COMPOSE_FILE) logs -f $(APP_NAME)

.PHONY: docker-clean
docker-clean: docker-stop
	docker-compose -f $(DOCKER_COMPOSE_FILE) down -v --rmi all
	docker system prune -f

# Database commands
.PHONY: redis-start
redis-start:
	docker run -d --name cart-redis -p 6379:6379 redis:7-alpine

.PHONY: redis-stop
redis-stop:
	docker stop cart-redis && docker rm cart-redis

.PHONY: redis-cli
redis-cli:
	docker exec -it cart-redis redis-cli

# Development workflow
.PHONY: dev-setup
dev-setup: deps tidy

.PHONY: dev-test
dev-test: fmt lint test

.PHONY: dev-run
dev-run: build run

# CI/CD
.PHONY: ci
ci: deps tidy fmt lint test

.PHONY: release
release: ci docker-build

# Security scan (requires gosec to be installed)
.PHONY: security
security:
	gosec ./...

.PHONY: install-security
install-security:
	$(GOGET) github.com/securecodewarrior/gosec/v2/cmd/gosec
