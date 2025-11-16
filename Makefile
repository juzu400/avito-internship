APP_NAME := avito-internship
CMD_PATH := ./cmd/server

.PHONY: run build test tidy lint docker-up docker-down

run:
	go run $(CMD_PATH)

build:
	go build -o bin/$(APP_NAME) $(CMD_PATH)

test:
	go test ./...

tidy:
	go mod tidy

lint:
	golangci-lint run ./...

docker-up:
	docker compose up --build app db

docker-down:
	docker compose down -v

test-db-up:
	docker compose up -d db_test

test-e2e: test-db-up
	go test -tags=integration ./test/e2e -v
	docker compose rm -sf db_test
