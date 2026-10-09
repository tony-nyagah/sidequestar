.PHONY: run generate test check build

run:
	go run ./cmd/server

generate:
	go tool sqlc generate

build:
	go build ./...

test:
	go test ./...

check: generate
	go fmt ./...
	go vet ./...
	go test ./...
