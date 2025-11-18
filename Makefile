.PHONY: fmt build run local-compose-up local-compose-down mocks unit-test

fmt:
	go fmt ./...

build:
	go build ./...

run: 
	go run cmd/server/main.go

local-compose-up:
	docker compose -f docker-compose.local.yml up -d

local-compose-down:
	docker compose -f docker-compose.local.yml down

mocks:
	mockgen -destination=internal/storage/mocks/storage.go -package=mocks github.com/cursed-ninja/internal-transfers-system/internal/storage Storage

unit-test:
	go test -v ./... -coverprofile=coverage.out