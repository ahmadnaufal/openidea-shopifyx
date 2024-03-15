.PHONY: all test build

all: deps test build

build:
	go mod tidy
	go build -o ./build/main ./cmd/main.go

compile-linux:
	GOOS=linux GOARCH=amd64 go build -o ./build/linux-amd64/main ./cmd/main.go

compile-darwin:
	GOOS=darwin GOARCH=arm64 go build -o ./build/darwin-arm64/main ./cmd/main.go

deps:
	go mod tidy

test:
	go test ./...

create-migration:
	migrate create -ext sql -dir db/migrations todo_new_migration

migrate:
	migrate -path ./db/migrations -database $(DB_CONN_URL) up

migrate-rollback:
	migrate -path ./db/migrations -database $(DB_CONN_URL) down 1
