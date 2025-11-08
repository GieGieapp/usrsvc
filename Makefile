run:
	APP_PORT=$${APP_PORT:-8080} go run ./cmd/api
test:
	go test ./... -count=1
fmt:
	go fmt ./...
migrate-up:
	migrate -path migrations -database $${PG_DSN} up
migrate-down:
	migrate -path migrations -database $${PG_DSN} down 1
