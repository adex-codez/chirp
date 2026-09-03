.PHONY: run migrate-up migrate-down sqlc

run:
	go run github.com/air-verse/air@v1.61.7

migrate-up:
	go run ./cmd/migrate up

migrate-down:
	go run ./cmd/migrate down

sqlc:
	go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.28.0 generate
