.PHONY: run build test lint client dev stop start test-provider

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
	$(MAKE) -j2 air client

air:
	cd backend && air

stop:
	powershell -ExecutionPolicy Bypass -Command "Get-NetTCPConnection -LocalPort 8080,5173 -State Listen -ErrorAction SilentlyContinue | ForEach-Object { Stop-Process -Id \$$_.OwningProcess -Force -ErrorAction SilentlyContinue }; Write-Host 'Stopped'"

start: stop dev

test-provider:
	cd backend && go run ./cmd/test-provider
