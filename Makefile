.PHONY: build run-all test lint vet migrate clean docker-build tidy

# Build all services
build:
	@echo "Building all services..."
	go build -o bin/gateway ./cmd/gateway
	go build -o bin/wallet ./cmd/wallet
	go build -o bin/engine ./cmd/engine
	go build -o bin/settlement ./cmd/settlement
	go build -o bin/games ./cmd/games
	go build -o bin/migrate ./cmd/migrate

tidy:
	go mod tidy

vet:
	go vet ./...

lint:
	golangci-lint run ./...

# Run all services locally (requires Docker for dependencies)
run-all: build
	@echo "Starting all services..."
	docker-compose up -d postgres redis nats
	./bin/gateway & ./bin/wallet & ./bin/engine & ./bin/settlement & ./bin/games &

# Run database migrations
migrate:
	@echo "Running database migrations..."
	go run ./cmd/migrate -dir ./migrations

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/
	docker-compose down

# Build Docker images for production
docker-build:
	docker build -t betting-platform/gateway:latest -f deployments/gateway/Dockerfile .
	docker build -t betting-platform/wallet:latest -f deployments/wallet/Dockerfile .
	docker build -t betting-platform/engine:latest -f deployments/engine/Dockerfile .
	docker build -t betting-platform/settlement:latest -f deployments/settlement/Dockerfile .
	docker build -t betting-platform/games:latest -f deployments/games/Dockerfile .

# Development setup
dev-setup:
	@echo "Setting up development environment..."
	docker-compose up -d
	sleep 5
	make migrate
