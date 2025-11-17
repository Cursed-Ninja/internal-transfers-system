.PHONY: fmt build run

fmt:
	go fmt ./...

build:
	go build ./...

run: 
	APP_ENV=local \
	go run main.go