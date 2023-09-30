test:
	go test -timeout 30s -coverprofile=coverage.out ./...

run-migration:
	go run database/migration/main.go

run:
	go run .