.PHONY: generate test test-integration test-e2e lint migrate-up run
generate:
	sqlc generate
test:
	go test ./...
test-integration:
	go test -tags=integration ./...
test-e2e:
	go test -tags=e2e ./test/e2e
lint:
	go vet ./...
	staticcheck ./...
migrate-up:
	goose -dir db/migrations postgres "$(DATABASE_URL)" up
run:
	go run ./cmd/api
