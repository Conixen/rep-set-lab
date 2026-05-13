.PHONY: run build test lint client dev

run:
	cd backend && go run ./cmd/api

build:
	cd backend && go build -o bin/api ./cmd/api

test:
	cd backend && go test ./...

lint:
	cd backend && golangci-lint run ./...

client:
	cd client && npm run dev

dev:
	$(MAKE) -j2 run client
