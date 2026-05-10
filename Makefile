.PHONY: run build migrate seed test lint

run:
	go run ./cmd/api

build:
	go build -o bin/api ./cmd/api

migrate:
	@echo "Migrations run automatically on startup via database.Migrate()"

seed:
	@echo "Exercise seed runs automatically on startup via exerciseStore.Seed()"

test:
	go test ./...

lint:
	golangci-lint run ./...

tidy:
	go mod tidy
