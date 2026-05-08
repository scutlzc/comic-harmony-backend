.PHONY: build run migrate lint clean docker-up docker-down docker-rebuild

APP=comic-harmony-backend
CMD=./cmd/server

build:
	go build -o bin/$(APP) $(CMD)

run: build
	./bin/$(APP)

dev:
	go run $(CMD)/main.go

migrate:
	psql "$(DATABASE_URL)" -f migrations/001_init.sql

lint:
	go vet ./...
	golangci-lint run ./... 2>/dev/null || true

test:
	go test ./... -v -count=1

clean:
	rm -rf bin/

docker-build:
	docker build -t $(APP) .

docker-up:
	docker compose up -d

docker-down:
	docker compose down -v

docker-rebuild:
	docker compose build --no-cache && docker compose up -d

docker-logs:
	docker compose logs -f

.PHONY: all
all: lint test build
