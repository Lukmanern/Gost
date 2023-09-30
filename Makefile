test:
	go test -timeout 30s -coverprofile=coverage.out ./...

migrate:
	go run database/migration/main.go

run:
	go run .