.PHONY: build run test clean docker-up docker-down migrate

# Build the application
build:
	go build -o bin/urlshortener ./cmd/server

# Run the application
run:
	go run ./cmd/server

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Start development dependencies (PostgreSQL and Redis)
docker-up:
	docker-compose up -d postgres redis

# Stop development dependencies
docker-down:
	docker-compose down

# Download dependencies
deps:
	go mod tidy
	go mod download

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	golangci-lint run

# Build Docker image
docker-build:
	docker build -t urlshortener .

# Run the full application stack
docker-run-all:
	docker-compose up

# Development setup (start dependencies and run app)
dev: docker-up
	@echo "Waiting for database to be ready..."
	@sleep 3
	@echo "Starting application..."
	go run ./cmd/server 