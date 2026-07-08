.PHONY: build run test linter tidy

build:
	go build -o bin/api ./cmd/api

run:
	go run ./cmd/api

test:
	go test -race -cover ./...

linter:
	golangci-lint run ./...

tidy:
	go mod tidy
