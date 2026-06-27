.PHONY: help build run test docker-up docker-down docker-logs clean

help:
	@echo "Available commands:"
	@echo "  make build          - Build the application binary"
	@echo "  make run            - Run the application locally"
	@echo "  make test           - Run tests"
	@echo "  make docker-up      - Start Docker containers"
	@echo "  make docker-down    - Stop Docker containers"
	@echo "  make docker-logs    - View Docker logs"
	@echo "  make docker-rebuild - Rebuild Docker image"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make deps           - Download dependencies"

build:
	@echo "Building api-mobile..."
	go build -o bin/api-mobile ./cmd/main.go
	@echo "Build complete: bin/api-mobile"

run: build
	@echo "Running api-mobile..."
	./bin/api-mobile

test:
	@echo "Running tests..."
	go test -v -race ./...

docker-up:
	@echo "Starting Docker containers..."
	docker-compose up -d
	@echo "Containers started"
	@echo "API will be available at http://localhost:8080"

docker-down:
	@echo "Stopping Docker containers..."
	docker-compose down
	@echo "Containers stopped"

docker-logs:
	docker-compose logs -f api-mobile

docker-rebuild:
	@echo "Rebuilding Docker image..."
	docker-compose build --no-cache api-mobile
	docker-compose up -d api-mobile
	@echo "Image rebuilt and container restarted"

clean:
	@echo "Cleaning..."
	rm -f bin/api-mobile
	go clean
	@echo "Clean complete"

deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy
	@echo "Dependencies downloaded"

fmt:
	@echo "Formatting code..."
	go fmt ./...
	@echo "Format complete"

lint:
	@echo "Linting code..."
	golangci-lint run ./...

run-local:
	@echo "Running locally with local database (ensure PostgreSQL and Redis are running)"
	go run ./cmd/main.go
