.PHONY: help up down restart env-from-template certs deps swagdocs test lint

BUILD_COMMIT  ?= $(shell git rev-parse --short HEAD)
BUILD_DATE    ?= $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
BUILD_VERSION ?= $(shell git tag -l 'server-version-*' --sort=-version:refname | head -n1 | sed 's/server-version-//' || echo "N/A")

export BUILD_COMMIT BUILD_DATE BUILD_VERSION

## General
help:  ## List of available commands
	@echo "Available commands:"
	@awk 'BEGIN {FS = ":.*?## "}; /^## / {printf "\033[1;34m%s\033[0m\n", substr($$0, 4); next}; /^[a-zA-Z_-]+:.*?## / {printf "  \033[1;32m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

## Containers
up:  ## Build and start containers
	@echo "Start building with:"
	@echo "  BUILD_COMMIT=$(BUILD_COMMIT)"
	@echo "  BUILD_DATE=$(BUILD_DATE)"
	@echo "  BUILD_VERSION=$(BUILD_VERSION)"
	docker-compose build
	docker-compose up -d

down:  ## Stop and remove containers and networks
	docker-compose down -v
	docker volume ls -q | xargs -r docker volume rm
	docker system prune -af --volumes

restart: down up  ## Restart services

## Configuration
env-from-template:  ## Create .env file from template
	@if [ ! -f .env ]; then cp .env.template .env; fi

certs:  ## Generate self-signed TLS certificates
	@./scripts/generate_certs.sh

## Development
deps:  ## Update Go dependencies
	rm -f go.sum
	go mod tidy -v
	go mod verify

swagdocs:  ## Generate Swagger documentation
	swag init --dir ./cmd/server,./internal/server/delivery --output ./docs

test:  ## Run all tests with coverage analysis
	@echo "\n\033[1;34mRun Tests:\033[0m\n"
	@go test -v -coverprofile=coverage.out ./... 
	@echo "\n\033[1;34mCoverage Analysis:\033[0m\n"
	@go tool cover -func=coverage.out
	@rm -f coverage.out

lint:  ## Run linter, format code, and generate report
	-fieldalignment -fix ./... || true
	-goimports -w . || true
	-gofmt -w . || true
	-golines -m 110 -w . || true
	mkdir -p ./golangci-lint
	-golangci-lint run -c .golangci.yml --out-format json > ./golangci-lint/report-unformatted.json || true
	cat ./golangci-lint/report-unformatted.json | jq '{IssuesCount: (.Issues | length), Issues: [.Issues[] | {FromLinter, Text, SourceLines, Filename: .Pos.Filename, Line: .Pos.Line}]}' > ./golangci-lint/report.json
	rm ./golangci-lint/report-unformatted.json