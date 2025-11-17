.PHONY: fmt build run

fmt:
	go fmt ./...

build:
	go build ./...

run: 
	go run cmd/server/main.go

local-compose:
	docker compose -f docker-compose.local.yml up -d