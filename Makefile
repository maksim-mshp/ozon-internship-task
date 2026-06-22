.PHONY: install-deps up test test-integration lint openapi coverage start-docker

install-deps:
	@go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
	@go install github.com/swaggo/swag/v2/cmd/swag@latest

up:
	@go run ./cmd/shortener

test:
	@go test -v -short ./...

test-integration:
	@go test -v ./tests/integration/...

lint:
	@golangci-lint run ./...

openapi: install-deps
	@swag init -g internal/httpserver/swagger.go --exclude _codex --output api --outputTypes json,yaml --v3.1
	@mv api/swagger.json api/openapi.json
	@mv api/swagger.yaml api/openapi.yml

coverage:
	@go test -coverpkg=./... ./... -coverprofile=coverage/coverage.out
	@go tool cover -html=coverage/coverage.out -o coverage/coverage.html
	@printf 'Total coverage: '
	@go tool cover -func=coverage/coverage.out | awk '/^total:/ {print $$3}'

start-docker:
	@docker compose up --build
