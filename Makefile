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
	docker-compose up --build

docker-down:
	docker-compose down -v
