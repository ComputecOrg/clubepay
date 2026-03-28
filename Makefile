.PHONY: dev-backend dev-frontend test-backend test-frontend test migrate-up migrate-down sqlc docker-up docker-down docker-logs

# Development
dev-backend:
	cd backend && go run ./cmd/api

dev-frontend:
	cd frontend && npm run dev

# Tests
test-backend:
	cd backend && go test ./...

test-frontend:
	cd frontend && npx vitest run

test: test-backend test-frontend

# Database
migrate-up:
	cd backend && migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	cd backend && migrate -path migrations -database "$(DATABASE_URL)" down

# Code generation
sqlc:
	cd backend && sqlc generate

# Build
build-backend:
	cd backend && go build -o bin/api ./cmd/api

build-frontend:
	cd frontend && npm run build

build: build-backend build-frontend

# Docker
docker-backend:
	docker build -f Dockerfile.backend -t clubepay-api .

docker-frontend:
	docker build -f Dockerfile.frontend -t clubepay-web .

# Docker Compose
docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f
