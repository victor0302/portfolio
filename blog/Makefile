.PHONY: help run build test vet tidy clean seed post-list post-help

PORT ?= 8080
BLOG_DB ?= blog.db

help:
	@echo "available targets:"
	@echo "  run        — start the server (PORT=$(PORT), BLOG_DB=$(BLOG_DB))"
	@echo "  build      — compile bin/server, bin/seed, bin/post"
	@echo "  test       — run all tests"
	@echo "  vet        — run go vet"
	@echo "  tidy       — go mod tidy"
	@echo "  seed       — populate the db with sample posts (idempotent)"
	@echo "  post-list  — list all posts (published + drafts)"
	@echo "  post-help  — print the post CLI usage"
	@echo "  clean      — remove bin/"

run:
	PORT=$(PORT) BLOG_DB=$(BLOG_DB) go run ./cmd/server

build:
	go build -o bin/server ./cmd/server
	go build -o bin/seed   ./cmd/seed
	go build -o bin/post   ./cmd/post

test:
	go test ./...

vet:
	go vet ./...

tidy:
	go mod tidy

seed:
	go run ./cmd/seed -db $(BLOG_DB)

post-list:
	go run ./cmd/post list -db $(BLOG_DB) -all

post-help:
	go run ./cmd/post help

clean:
	rm -rf bin/
