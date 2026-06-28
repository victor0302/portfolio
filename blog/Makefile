.PHONY: help run build test tidy clean

PORT ?= 8080

help:
	@echo "available targets:"
	@echo "  run    — start the server (PORT=$(PORT))"
	@echo "  build  — compile to bin/server"
	@echo "  test   — run all tests"
	@echo "  tidy   — go mod tidy"
	@echo "  clean  — remove bin/"

run:
	PORT=$(PORT) go run ./cmd/server

build:
	go build -o bin/server ./cmd/server

test:
	go test ./...

tidy:
	go mod tidy

clean:
	rm -rf bin/
