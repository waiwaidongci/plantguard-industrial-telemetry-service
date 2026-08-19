.PHONY: tidy fmt test vet build run-dev run-api run-worker migrate-postgres count

tidy:
	go mod tidy

fmt:
	gofmt -w .

test:
	go test ./...

vet:
	go vet ./...

build:
	go build ./...

run-dev:
	./scripts/run-dev.sh

run-api:
	go run ./cmd/api

run-worker:
	go run ./cmd/worker

migrate-postgres:
	./scripts/migrate-postgres.sh

count:
	./scripts/count-go.sh
